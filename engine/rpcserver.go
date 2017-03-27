package engine

import (
	"net"
	"net/http"
	"time"

	log "github.com/funkygao/log4go"
	"github.com/gorilla/mux"
)

func (this *Engine) launchRPCServer() {
	this.rpcRouter = mux.NewRouter()
	this.rpcServer = &http.Server{
		Addr:         this.participantID,
		Handler:      this.rpcRouter,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	this.setupRPCRoutings()

	var err error
	if this.rpcListener, err = net.Listen("tcp", this.rpcServer.Addr); err != nil {
		panic(err)
	}

	go this.rpcServer.Serve(this.rpcListener)
	log.Info("RPC server ready on http://%s", this.rpcServer.Addr)
}

func (this *Engine) stopRPCServer() {
	if this.rpcListener != nil {
		this.rpcListener.Close()
		log.Info("RPC server stopped")
	}
}

func (this *Engine) setupRPCRoutings() {
	this.rpcRouter.HandleFunc("path", this.doRebalance)
}

func (this *Engine) doRebalance(w http.ResponseWriter, r *http.Request) {

}
