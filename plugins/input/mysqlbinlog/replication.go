package mysqlbinlog

import (
	"context"
	"fmt"

	conf "github.com/funkygao/jsconf"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
)

type MysqlBinlog struct {
	r *replication.BinlogSyncer
}

func New() *MysqlBinlog {
	return &MysqlBinlog{}
}

func (m *MysqlBinlog) LoadConfig(config *conf.Conf) *MysqlBinlog {
	m.r = replication.NewBinlogSyncer(&replication.BinlogSyncerConfig{
		ServerID: uint32(config.Int("server_id", 1007)),
		Flavor:   config.String("flavor", "mysql"), // or mariadb
		Host:     config.String("master_host", "localhost"),
		Port:     uint16(config.Int("master_port", 3306)),
		User:     config.String("user", "root"),
		Password: config.String("password", ""),
	})
	return m
}

func (m *MysqlBinlog) Start() error {
	pos := mysql.Position{}
	s, err := m.r.StartSync(pos)
	if err != nil {
		panic(err)
	}

	for {
		ctx := context.Background()
		ev, err := s.GetEvent(ctx)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// next binglog pos
		//pos = ev.Header.LogPos

		fmt.Printf("%#v\n", ev)
	}

	return nil
}

func (m *MysqlBinlog) Close() {

}
