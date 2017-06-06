package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gocli"
	"github.com/github/gh-ost/go/base"
	"github.com/github/gh-ost/go/logic"
	"github.com/outbrain/golib/log"
)

type Migrate struct {
	Ui  cli.Ui
	Cmd string
}

// record history of all migrates
func (this *Migrate) Run(args []string) (exitCode int) {
	var how bool
	ctx := base.GetMigrationContext()
	cmdFlags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	cmdFlags.StringVar(&ctx.InspectorConnectionConfig.Key.Hostname, "host", "127.0.0.1", "")
	cmdFlags.IntVar(&ctx.InspectorConnectionConfig.Key.Port, "port", 3306, "")
	cmdFlags.StringVar(&ctx.DatabaseName, "db", "", "")
	cmdFlags.StringVar(&ctx.OriginalTableName, "table", "", "")
	cmdFlags.StringVar(&ctx.AlterStatement, "alter", "", "")
	cmdFlags.BoolVar(&how, "how", false, "")
	cmdFlags.BoolVar(&ctx.Noop, "dryrun", false, "")
	cmdFlags.StringVar(&ctx.CliUser, "user", "root", "")
	cmdFlags.StringVar(&ctx.CliPassword, "pass", "", "")
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if how {
		this.printHowItWorks()
		return
	}

	ctx.AllowedRunningOnMaster = true
	ctx.OkToDropTable = true
	ctx.InitiallyDropOldTable = true
	ctx.InitiallyDropGhostTable = true
	ctx.ServeSocketFile = "/tmp/dbc.migrate.sock"
	ctx.SetChunkSize(1000)
	ctx.SetDMLBatchSize(100)
	ctx.SetDefaultNumRetries(3)
	ctx.ApplyCredentials()

	log.SetLevel(log.DEBUG)

	log.Infof("starting")

	// a migrator contains a sql parser(parsing alter statement), an inspector, applier and binlog streamer.
	// migrator.onChangelogStateEvent
	// migrator.onApplyEventStruct
	// 1. change log table
	// 2. ghost table
	// 3. alter ghost table
	migrator := logic.NewMigrator()
	if err := migrator.Migrate(); err != nil {
		migrator.ExecOnFailureHook()
		log.Fatale(err)
	}
	log.Infof("OK %d rows copied, %s, %s", ctx.GetTotalRowsCopied(), ctx.ElapsedRowCopyTime(), ctx.ElapsedTime())

	return
}

func (this *Migrate) printHowItWorks() {
	this.Ui.Output(strings.TrimSpace(`
1.parse alter statement

2.init inspector
    validate connection
        select @@global.port, @@global.version
    validate grants
        show /* gh-ost */ grants for current_user()
    validate binlogs
        select @@global.log_bin, @@global.binlog_format
    validate original table, ensure it exists and is not a VIEW
        show /* gh-ost */ table status from %s like '%s'
        make sure no foreign keys, no triggers
        estimate num of rows
            explain select /* gh-ost */ * from %s.%s where 1=1
    inspect original table columns and uniq keys

3.init binlog streaming
    read current binlog coordinates
        show /* gh-ost readCurrentBinlogCoordinates */ master status
    start binlog replication
    change log table -> onChangelogStateEvent
        GhostTableMigrated
        AllEventsUpToLockProcessed 
   
4.init applier
    open 2 conns to mysql
    select @@global.time_zone
    get original table columns
        show columns from %s.%s
    CREATE CHANGELOG TABLE
    CREATE GHOST TABLE
        create /* gh-ost */ table %s.%s like %s.%s
    ALTER GHOST TABLE
    WriteChangelogState(GhostTableMigrated)
    init heartbeat

5.inspect replication lag
    show slave status // Seconds_Behind_Master

6.inspectOriginalAndGhostTables
    find shared uniq keys

7.count original table rows
    select /* gh-ost */ count(*) as rows from %s.%s

8.ADD DML Event Listener
    original table   -> applyEventsQueue

9.read migration range min/max
    select /* gh-ost %s.%s */ uniqKeyColumnsCommaSep from %.%s order by limit 1

10.init throttler

11.executeWriteFuncs
    write data via applier: rowcopy and applyEventsQueue
    applier.ApplyDMLEventQueries
        INSERT: replace into %s.%s (%s) values(%s)
        UPDATE: update %s.%s set %s where %s
        DELETE: delete from %s.%s where %s
    call copyRowsFunc

12.iterateChunks
    calculate chunk range

    MigrationIterationRangeMaxValues=select id from (select id from a where id>=0 and id<=3001 order by id asc limit 1000) select_osc_chunk order by id desc limit 1;

13.wait till row copy complete

14.retry cut over
    conn1: CREATE TABLE tbl_old (id int primary key) COMMENT='magic-be-here' // sentry table
    conn1: LOCK TABLES tbl WRITE, tbl_old WRITE
    conn2: RENAME TABLE tbl TO tbl_old, ghost TO tbl // blocked due to lock, but prioritized on top of other DML conns
    conn1: checks that conn2's RENAME is applied(show processlist)
    conn1: DROP TABLE tbl_old // ok for a conn to drop a table it has under write lock
    conn1: UNLOCK TABLES
    conn2: RENAME is the first to execute, then other DML conns, and the DML will operate on the new tbl

    // a blocked RENAME is always prioritized over a blocked INSERT/DELETE/UPDATE, no matter who came first

    select connection_id()             // get session id
    select get_lock(gh-ost.%d.lock, 0) // try lock
    set session lock_wait_timeout:=2
    create cut over sentry table _{tbl}_del
    lock /* gh-ost */ tables {tbl} write, _{tbl}_del write

    waitForEventsUpToLock
    rename
    unlock tables

15.final cleanup
    drop change log table
    drop old table
    drop ghost table
`))
}

func (*Migrate) Synopsis() string {
	return "Asynchronous MySQL online schema migration tool"
}

func (this *Migrate) Help() string {
	help := fmt.Sprintf(`
Usage: %s migrate [options]

    %s

Options:

    -user db username

    -pass db password

    -db dabase

    -table table

    -alter statement
      e,g.
      dbc migrate -db a -table a -alter 'add age int'

    -dryrun

    -status

    -how

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
