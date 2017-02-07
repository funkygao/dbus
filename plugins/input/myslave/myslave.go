package myslave

import (
	"fmt"

	conf "github.com/funkygao/jsconf"
	"github.com/siddontang/go-mysql/replication"
)

type MySlave struct {
	r *replication.BinlogSyncer

	gtid       bool // TODO
	masterAddr string
	pos        *checkpoint
}

func New() *MySlave {
	return &MySlave{}
}

func (m *MySlave) LoadConfig(config *conf.Conf) *MySlave {
	m.gtid = config.Bool("gtid", false)
	host := config.String("master_host", "localhost")
	port := uint16(config.Int("master_port", 3306))
	m.masterAddr = fmt.Sprintf("%s:%d", host, port)
	m.r = replication.NewBinlogSyncer(&replication.BinlogSyncerConfig{
		ServerID: uint32(config.Int("server_id", 1007)),
		Flavor:   config.String("flavor", "mysql"), // or mariadb
		Host:     host,
		Port:     port,
		User:     config.String("user", "root"),
		Password: config.String("password", ""),
	})
	return m
}

func (m *MySlave) MarkAsProcessed(name string, pos uint32) error {
	m.pos.Update(name, pos)
	return m.pos.Save(false)
}
