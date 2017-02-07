package mysqlbinlog

import (
	conf "github.com/funkygao/jsconf"
	"github.com/siddontang/go-mysql/canal"
)

type MysqlBinlog struct {
	*canal.Canal
}

func New() *MysqlBinlog {
	return &MysqlBinlog{}
}

func (m *MysqlBinlog) LoadConfig(config *conf.Conf) *MysqlBinlog {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = config.String("master_addr", "localhost:3306")
	cfg.User = config.String("user", "root")
	cfg.Password = config.String("password", "")
	cfg.Flavor = config.String("flavor", "mysql") // or mariadb
	cfg.ServerID = uint32(config.Int("server_id", 1007))
	cfg.DataDir = config.String("work_dir", "var/")
	cfg.Dump.DiscardErr = false
	cfg.Dump.ExecutionPath = "" // ignore mysqldump
	// TODO validate

	c, err := canal.NewCanal(cfg)
	if err != nil {
		panic(err)
	}

	if image := config.String("binlog_row_image", ""); image != "" {
		if err = c.CheckBinlogRowImage(image); err != nil {
			panic(err)
		}
	}

	m.Canal = c
	return m
}
