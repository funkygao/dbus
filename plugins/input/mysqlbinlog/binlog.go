package mysqlbinlog

type MysqlBinlog struct {
}

func New() *MysqlBinlog {
	return &MysqlBinlog{}
}

func (b *MysqlBinlog) Stream() <-chan []byte {
	return nil
}
