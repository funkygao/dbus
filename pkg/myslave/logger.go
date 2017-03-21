package myslave

import (
	"github.com/ngaut/log"
)

func init() {
	// github.com/siddontang/go-mysql is using github.com/ngaut/log
	log.SetLevel(log.LOG_LEVEL_ERROR)
	if err := log.SetOutputByName("myslave.log"); err != nil {
		panic(err)
	}
}
