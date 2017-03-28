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

func (p *Project) fromConfig(c *conf.Conf) {
	p.Name = c.String("name", "")
	if p.Name == "" {
		panic("project must has 'name'")
	}
	p.ShowError = c.Bool("show_error", true)

	logfile := c.String("logfile", "var/"+p.Name+".log")
	logWriter, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	logOptions := log.Ldate | log.Ltime
	logOptions |= log.Lshortfile
	if Globals().Debug {
		logOptions |= log.Lmicroseconds
	}

	p.Logger = log.New(logWriter, "", logOptions)
}

func (p *Project) Start() {
	p.Printf("Project[%s] started", p.Name)
}

func (p *Project) Stop() {
	p.Printf("Project[%s] stopped", p.Name)
}
