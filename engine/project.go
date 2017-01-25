package engine

import (
	"log"
	"os"

	conf "github.com/funkygao/jsconf"
)

type ConfProject struct {
	*log.Logger

	Name        string `json:"name"`
	IndexPrefix string `json:"index_prefix"`
	ShowError   bool   `json:"show_error"`
}

func (this *ConfProject) fromConfig(c *conf.Conf) {
	this.Name = c.String("name", "")
	if this.Name == "" {
		panic("project must has 'name'")
	}
	this.IndexPrefix = c.String("index_prefix", this.Name)
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

func (this *ConfProject) Start() {
	this.Println("Started")
}

func (this *ConfProject) Stop() {
	this.Println("Stopped")
}
