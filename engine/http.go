package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/funkygao/golib/bjtime"
	"github.com/gorilla/mux"
)

func (this *Engine) launchHttpServ() {
	this.httpRouter = mux.NewRouter()
	this.httpServer = &http.Server{Addr: this.String("http_addr", "127.0.0.1:9876"), Handler: this.httpRouter}

	this.RegisterHttpApi("/admin/{cmd}",
		func(w http.ResponseWriter, req *http.Request,
			params map[string]interface{}) (interface{}, error) {
			return this.handleHttpQuery(w, req, params)
		}).Methods("GET")

	var err error
	this.httpListener, err = net.Listen("tcp", this.httpServer.Addr)
	if err != nil {
		panic(err)
	}
	go this.httpServer.Serve(this.httpListener)

	Globals().Printf("Listening on http://%s", this.httpServer.Addr)
}

func (this *Engine) handleHttpQuery(w http.ResponseWriter, req *http.Request,
	params map[string]interface{}) (interface{}, error) {
	var (
		vars    = mux.Vars(req)
		cmd     = vars["cmd"]
		globals = Globals()
		output  = make(map[string]interface{})
	)

	switch cmd {
	case "ping":
		output["status"] = "ok"

	case "shutdown":
		globals.Shutdown()
		output["status"] = "ok"

	case "reload", "restart":
		break

	case "debug":
		stack := make([]byte, 1<<20)
		stackSize := runtime.Stack(stack, true)
		globals.Println(string(stack[:stackSize]))
		output["result"] = "go to global logger to see result"

	case "stat":
		output["runtime"] = this.stats.Runtime()
		output["router"] = this.router.stats
		output["started"] = globals.StartedAt
		output["elapsed"] = time.Since(globals.StartedAt).String()
		output["pid"] = this.pid
		output["hostname"] = this.hostname

	case "pools":
		for poolName, _ := range this.diagnosticTrackers {
			packs := make([]string, 0, globals.RecyclePoolSize)
			for _, pack := range this.diagnosticTrackers[poolName].packs {
				s := fmt.Sprintf("[%s]%s",
					bjtime.TimeToString(pack.diagnostics.LastAccess),
					*pack)
				packs = append(packs, s)
			}
			output[poolName] = packs
			output[poolName+"_len"] = len(packs)
		}

	case "plugins":
		output["plugins"] = this.pluginNames()

	case "uris":
		output["all"] = this.httpPaths

	default:
		return nil, errors.New("Not Found")
	}

	return output, nil
}

func (this *Engine) RegisterHttpApi(path string,
	handlerFunc func(http.ResponseWriter,
		*http.Request, map[string]interface{}) (interface{}, error)) *mux.Route {
	wrappedFunc := func(w http.ResponseWriter, req *http.Request) {
		var (
			ret     interface{}
			globals = Globals()
			t1      = time.Now()
		)

		params, err := this.decodeHttpParams(w, req)
		if err == nil {
			ret, err = handlerFunc(w, req, params)
		}

		if err != nil {
			ret = map[string]interface{}{"error": err.Error()}
		}

		w.Header().Set("Content-Type", "application/json")
		var status int
		if err == nil {
			status = http.StatusOK
		} else {
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)

		// debug request body content
		if globals.Debug {
			globals.Printf("req body: %+v", params)
		}
		// access log
		globals.Printf("%s \"%s %s %s\" %d %s",
			req.RemoteAddr,
			req.Method,
			req.RequestURI,
			req.Proto,
			status,
			time.Since(t1))
		if status != http.StatusOK {
			globals.Printf("ERROR %v", err)
		}

		if ret != nil {
			// pretty write json result
			pretty, _ := json.MarshalIndent(ret, "", "    ")
			w.Write(pretty)
			w.Write([]byte("\n"))
		}
	}

	// path can't be duplicated
	for _, p := range this.httpPaths {
		if p == path {
			panic(path + " already registered")
		}
	}

	this.httpPaths = append(this.httpPaths, path)
	return this.httpRouter.HandleFunc(path, wrappedFunc)
}

func (this *Engine) decodeHttpParams(w http.ResponseWriter, req *http.Request) (map[string]interface{},
	error) {
	params := make(map[string]interface{})
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return params, nil
}

func (this *Engine) stopHttpServ() {
	if this.httpListener != nil {
		this.httpListener.Close()

		globals := Globals()
		if globals.Verbose {
			globals.Println("HTTP stopped")
		}
	}
}
