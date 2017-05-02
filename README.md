# dbus
    
          $$\       $$\                                       
          $$ |      $$ |                                      
     $$$$$$$ |      $$$$$$$\        $$\   $$\        $$$$$$$\ 
    $$  __$$ |      $$  __$$\       $$ |  $$ |      $$  _____|
    $$ /  $$ |      $$ |  $$ |      $$ |  $$ |      \$$$$$$\  
    $$ |  $$ |      $$ |  $$ |      $$ |  $$ |       \____$$\ 
    \$$$$$$$ |      $$$$$$$  |      \$$$$$$  |      $$$$$$$  |
     \_______|      \_______/        \______/       \_______/ 

### What is dbus?

dbus = distributed data bus

It is yet another lightweight versatile databus system that transfer/transform pipeline data between plugins.

dbus works by building a DAG of structured data out of the different plugins: from data input, via filter(optional), to the output.

Similar projects

- logstash
- flume
- nifi
- camel
- beats
- kettle
- zapier
- google cloud dataflow
- canal
- storm
- yahoo pipes (dead)

### Status

dbus is not yet a 1.0.
We're writing more tests, fixing bugs, working on TODOs.

### Use Case

- mysql binlog dispatcher
- multiple DC kafka mirror

### Features

dbus supports powerful and scalable directed graphs of data routing, transformation and
system mediation logic.

- Designed for extension
  - plugin architecture
  - build your own plugins and more
  - enables rapid development and effective testing
- Data Provenance
  - track dataflow from beginning to end
  - visualized dataflow
  - rich metrics feed into tsdb
  - online manual mediation of the dataflow
  - RESTful API
  - monitoring with alert
- Distributed Deployment
  - shard/balance/auto rebalance
  - linear scale
- Delivery Guarantee
  - loss tolerant
  - high throuput vs low latency
  - back pressure
- Robustness
  - race condition detected
  - edge cases fully covered
  - network jitter tested
  - dependent components failure tested
- Systemic Quality
  - hot reload
  - dryrun throughput 1.9M packets/s
- Cluster Support
  - modelling borrowed from helix+kafka controller
  - currently only leader/standby with sharding, without replica
  - easy to write a distributed plugin

### Getting Started

#### 1. Installing

To start using dbus, install Go and run `go get`:

```sh
$ go get -u github.com/funkygao/dbus
```

#### 2. Create config file

Please find sample config files in etc/ directory.

#### 3. Run the server

```sh
$ $GOPATH/dbusd -conf $myfile
```

### Dependencies

dbus uses zookeeper for sharding/balance/election.

### Plugins

More plugins are listed under [dbus-plugin](https://github.com/dbus-plugin).

#### Input

- MysqlbinlogInput
- KafkaInput
- MockInput
- StreamInput
  
#### Filter

- MysqlbinlogFilter
- MockFilter

#### Output

- KafkaOutput
- ESOutput
- MockOutput
- StreamOutput

### Configuration

- KafkaOutput async mode with batch=1024/500ms, ack=WaitForAll
- KafkaOutput retry=3, retry.backoff=350ms
- Mysql binlog positioner commit every 1s, channal buffer 100

### FAQ

#### mysql binlog

- is it totally data loss tolerant?
  
  if the binlog exceeds 1MB, it will be discarded(lost)

- ERROR 1236 (HY000): Could not find first log file name in binary log index file

  the checkpointed binlog position is gone on master, reset the zk znode and replication will 
  automatically resume

- how to migrate mysql database?

  mysqldump --user=root --master-data --single-transaction --skip-lock-tables --compact --skip-opt --quick --no-create-info --skip-extended-insert --all-databases

  parse CHANGE MASTER TO MASTER_LOG_FILE='mysql.000005', MASTER_LOG_POS=80955497;

  MysqlbinlogInput load position for incremental loading

#### why not canal?

- no Delivery Guarantee
- no Data Provenance
- no integration with kafka
- only hot standby deployment mode, we need sharding load
- dbus is a dataflow engine, while canal only support mysql binlog pipeline

#### compared with logstash

- logstash has better ecosystem
- dbus is cluster aware, provides delivery guarantee, data provenance

#### can there be more than 1 leaders at the same time?

Yes.

For example, 3 participants with 1 being the leader. Then 1 is network partitioned and
zk session expires, [2, 3] found this event and re-elect 2 as new leader.
Before 1 regain new zk session, [1] and [2] are leaders both.
If [1] and [2] both found resources changes, they will both rebalance the cluster.

dbus uses epoch to solve this issue.

#### what if

- zookeeper crash

  dbus continues to work, but Ack will not be able to persist

### TODO

- [ ] model.RowsEvent add dbus timestamp
- [ ] cluster
  - [ ] monitor resources cost and rebalance
  - [ ] support multiple projects
- [ ] resource group
- [ ] FIXME access denied leads to orphan resource
- [ ] myslave should have no checkpoint, placed in Input
- [ ] enhance Decision.Equals to avoid thundering herd
- [ ] myslave server_id uniq across the cluster
- [ ] tweak of batcher yield
- [ ] add Operator for Filter
  - count, filter, regex, sort, split, rename
- [ ] RowsEvent avro
- [ ] controller
  - [ ] a participant is electing, then shutdown took a long time(blocked by CreateLiveNode)
  - [X] 2 phase rebalance: close participants then notify new resources
  - [X] what if RPC fails
  - [X] leader.onBecomingLeader is parallal: should be sequential
  - [ ] hot reload raises cluster herd: participant changes too much
  - [X] when leader make decision, it persists to zk before RPC for leader failover
  - [X] owner of resource
  - [X] leader RPC has epoch info
  - [ ] if Ack fails(zk crash), resort to local disk(load on startup)
  - [X] engine shutdown, controller still send rpc
  - test cases
    - [X] sharded resources
    - [X] brain split
    - [X] zk dies or kill -9, use cache to continue work
    - [X] kill -9 participant/leader, and reschedule
    - [X] cluster chaos monkey
- [ ] kafka producer qos
- [ ] batcher only retries after full batch ack'ed, add timer?
- [ ] KafkaConsumer might not be able to Stop
- [ ] pack.Payload reuse memory, json.NewEncoder(os.Stdout)
- [X] kguard integration
- [X] router finding matcher is slow
- [X] hot reload on config file changed
- [X] each Input have its own recycle chan, one block will not block others
- [X] when Input stops, Output might still need its OnAck
- [X] KafkaInput plugin
- [X] use scheme to distinguish type of DSN
- [X] plugins Run has no way of panic
- [X] (replication.go:117) [zabbix] invalid table id 2968, no correspond table map event
- [X] make canal, high cpu usage
  - because CAS backoff 1us, cpu busy
- [X] ugly design of Input/Output ack mechanism
  - we might learn from storm bolt ack
- [X] some goroutine leakage
- [X] telemetry mysql.binlog.lag/tps tag name should be input name
- [X] pipeline
  - 1 input, multiple output
  - filter to dispatch dbs of a single binlog to different output
- [X] kill Packet.input field
- [X] visualized flow throughput like nifi
  - dump uses dag pkg
  - ![pipeline](https://github.com/funkygao/dbus-extra/blob/master/assets/dag.png?raw=true)
- [X] router metrics
- [X] dbusd api server
- [X] logging
- [X] share zkzone instance
- [X] presence and standby mode
- [X] graceful shutdown
- [X] master must drain before leave cluster
- [X] KafkaOutput metrics
  -  binlog tps
  -  kafka tps
  -  lag
- [X] hub is shared, what if a plugin blocks others
  - currently, I have no idea how to solve this issue
- [X] Batcher padding
- [X] shutdown kafka
- [X] zk checkpoint vs kafka checkpoint
- [X] kafka follower stops replication
- [X] can a mysql instance with miltiple databases have multiple Log/Position?
- [X] kafka sync produce in batch
- [X] DDL binlog
  - drop table y;
- [X] trace async producer Successes channel and mark as processed
- [X] metrics
- [X] telemetry and alert
- [X] what if replication conn broken
- [X] position will be stored in zk
- [X] play with binlog_row_image
- [ ] project feature for multi-tenant
- [ ] bug fix
  - [X] kill dbusd, dbusd-slave did not leave cluster
  - [X] next log position leads to failure after resume
  - [ ] KafkaOutput only support 1 partition topic for MysqlbinlogInput
  - [X] table id issue
  - [X] what if invalid position
  - [X] router stat wrong
    Total:142,535,625      0.00B speed:22,671/s      0.00B/s max: 0.00B/0.00B
  - [X] ffjson marshalled bytes has NL before the ending bracket
- [ ] test cases
  - [X] restart mysql master
  - [X] mysql kill process
  - [X] race detection
  - [ ] tc drop network packets and high latency
  - [ ] mysql binlog zk session expire
  - [X] reset binlog pos, and check kafka didn't recv dup events
  - [X] MysqlbinlogInput max_event_length
  - [X] min.insync.replicas=2, shutdown 1 kafka broker then start
- [ ] GTID
  - place config to central zk znode and watch changes
- [ ] Known issues
  - Binlog Dump thread not close https://github.com/github/gh-ost/issues/292
- [ ] Roadmap
  - pubsub audit reporter
  - universal kafka listener and outputer

### Memo

- mysqlbinlog input peak with mock output
  - 140k event per second
  - 30k row event per second
  - 260Mb network bandwidth
  - KafkaOutput 35K msg per second
  - it takes 2h25m to zero lag for platform of 2d lag

- dryrun MockInput -> MockOutput
  - 2.1M packet/s

