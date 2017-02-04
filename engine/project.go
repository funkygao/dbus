package engine

import (
	"log"
	"os"

	conf "github.com/funkygao/jsconf"
)

type Project struct {
	*log.Logger

	Name      string `json:"name"`
	ShowError bool   `json:"show_error"`
}

func (this *Project) fromConfig(c *conf.Conf) {
	this.Name = c.String("name", "")
	if this.Name == "" {
		panic("project must has 'name'")
	}
	this.ShowError = c.Bool("show_error", true)

	logfile := c.String("logfile", "var/"+this.Name+".log")
	logWriter, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	logOptions := log.Ldate | log.Ltime
	if Globals().Verbose {
		logOptions |= log.Lshortfile
	}
	if Globals().Debug {
		logOptions |= log.Lmicroseconds
	}

	this.Logger = log.New(logWriter, "", logOptions)
}

func (this *Project) Start() {
	this.Printf("Project[%s] started", this.Name)
}

func (this *Project) Stop() {
	this.Printf("Project[%s] stopped", this.Name)
}
