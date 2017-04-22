package http

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/middleware"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
	"github.com/gorilla/mux"
)

// HTTPInput is an input plugin that receives HTTP POST payload.
// For cluster, there should be a load balancer sitting in front of dbusd.
type HTTPInput struct {
	ex engine.Exchange
}

func (this *HTTPInput) Init(config *conf.Conf) {
}

func (*HTTPInput) SampleConfig() string {
	return `
	listen: "localhost:8899"
	`
}

func (this *HTTPInput) Ack(pack *engine.Packet) error {
	return nil
}

func (this *HTTPInput) End(r engine.InputRunner) {}

func (this *HTTPInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	addr := r.Conf().String("listen", "")
	if len(addr) == 0 {
		return errors.New("empty listen")
	}

	router := mux.NewRouter()
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	router.Handle("/v1/payload",
		middleware.WrapAccesslog(
			middleware.WrapWithRecover(
				http.HandlerFunc(this.payloadHandler)))).
		Methods("POST")

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	log.Trace("[%s] ready on %s", r.Name(), addr)
	this.ex = r.Exchange()
	go func(l net.Listener) {
		server.Serve(l)
	}(l)

	<-r.Stopper()

	return nil
}
