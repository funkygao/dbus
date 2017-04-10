package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/funkygao/dbus"
	"github.com/funkygao/golib/dag"
	log "github.com/funkygao/log4go"
	"github.com/gorilla/mux"
)

// APIHandler is the HTTP API server handler signature.
type APIHandler func(w http.ResponseWriter, req *http.Request, params map[string]interface{}) (interface{}, error)

func (e *Engine) launchAPIServer() {
	e.apiRouter = mux.NewRouter()
	e.apiServer = &http.Server{
		Addr:         strings.TrimLeft(e.participant.APIEndpoint(), "http://"),
		Handler:      e.apiRouter,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	e.setupAPIRoutings()

	var err error
	if e.apiListener, err = net.Listen("tcp", e.apiServer.Addr); err != nil {
		panic(err)
	}

	go e.apiServer.Serve(e.apiListener)
	log.Info("API server ready on http://%s", e.apiServer.Addr)
}

func (e *Engine) stopAPIServer() {
	if e.apiListener != nil {
		e.apiListener.Close()
		log.Info("API server stopped")
	}
}

func (e *Engine) setupAPIRoutings() {
	// admin
	e.RegisterAPI("/stat", e.handleAPIStat).Methods("GET")
	e.RegisterAPI("/plugins", e.handleAPIPlugins).Methods("GET")
	e.RegisterAPI("/metrics", e.handleAPIMetrics).Methods("GET")
	e.RegisterAPI("/dag", e.handleAPIDag).Methods("GET")

	// API
	e.RegisterAPI("/api/v1/pause/{input}", e.handleAPIPauseV1).Methods("PUT")
	e.RegisterAPI("/api/v1/resume/{input}", e.handleAPIResumeV1).Methods("PUT")
	e.RegisterAPI("/api/v1/decision", e.handleAPIDecisionV1).Methods("GET")
}

func (e *Engine) handleAPIDag(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
	d := dag.New()

	// vertex
	for in := range e.InputRunners {
		if _, present := e.router.metrics.m[in]; present {
			d.AddVertex(in, int(e.router.metrics.m[in].Rate1()))
		}
	}
	for _, m := range e.router.filterMatchers {
		if _, present := e.router.metrics.m[m.runner.Name()]; present {
			d.AddVertex(m.runner.Name(), int(e.router.metrics.m[m.runner.Name()].Rate1()))
		}
	}
	for _, m := range e.router.outputMatchers {
		if _, present := e.router.metrics.m[m.runner.Name()]; present {
			d.AddVertex(m.runner.Name(), int(e.router.metrics.m[m.runner.Name()].Rate1()))
		}
	}

	// edge
	for _, m := range e.router.filterMatchers {
		for source := range m.matches {
			d.AddEdge(source, m.runner.Name())
		}
	}
	for _, m := range e.router.outputMatchers {
		for source := range m.matches {
			log.Info("%s,%s %+v", source, m.runner.Name(), d)
			d.AddEdge(source, m.runner.Name())
		}
	}

	dir := os.TempDir()
	pngFile := fmt.Sprintf("%s/dag.png", dir)
	dot := d.MakeDotGraph(fmt.Sprintf("%s/dag.dot", dir))

	// the cmdLine is internal generated, should not vulnerable to security attack
	cmdLine := fmt.Sprintf("dot -o%s -Tpng -s3", pngFile)
	cmd := exec.Command(`/bin/sh`, `-c`, cmdLine)
	cmd.Stdin = strings.NewReader(dot)
	cmd.Run()

	w.Header().Set("Content-type", "image/png")
	b, _ := ioutil.ReadFile(pngFile)
	w.Write(b)

	// TODO schedule to delete the tmp files

	return nil, nil
}

func (e *Engine) handleAPIMetrics(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
	output := make(map[string]map[string]interface{})
	for ident, m := range e.router.metrics.m {
		output[ident] = map[string]interface{}{
			"tps": int(m.Rate1()),
			"cum": m.Count(),
		}
	}

	return output, nil
}

func (e *Engine) handleAPIPlugins(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
	plugins := make(map[string][]string)
	for _, r := range e.InputRunners {
		if _, present := plugins[r.Class()]; !present {
			plugins[r.Class()] = []string{r.Name()}
		} else {
			plugins[r.Class()] = append(plugins[r.Class()], r.Name())
		}
	}
	for _, r := range e.FilterRunners {
		if _, present := plugins[r.Class()]; !present {
			plugins[r.Class()] = []string{r.Name()}
		} else {
			plugins[r.Class()] = append(plugins[r.Class()], r.Name())
		}
	}
	for _, r := range e.OutputRunners {
		if _, present := plugins[r.Class()]; !present {
			plugins[r.Class()] = []string{r.Name()}
		} else {
			plugins[r.Class()] = append(plugins[r.Class()], r.Name())
		}
	}

	return plugins, nil
}

func (e *Engine) handleAPIStat(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
	var output = make(map[string]interface{})
	output["ver"] = dbus.Version
	output["started"] = Globals().StartedAt
	output["elapsed"] = time.Since(Globals().StartedAt).String()
	output["pid"] = e.pid
	output["hostname"] = e.hostname
	output["revision"] = dbus.Revision
	return output, nil
}

func (e *Engine) RegisterAPI(path string, handlerFunc APIHandler) *mux.Route {
	return e.apiRouter.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var (
			ret interface{}
			t1  = time.Now()
		)

		params, err := e.decodeHTTPParams(w, r)
		if err == nil {
			ret, err = handlerFunc(w, r, params)
		}

		if err != nil {
			ret = map[string]interface{}{"error": err.Error()}
		}

		w.Header().Set("Server", "dbus")
		var status int
		if err == nil {
			status = http.StatusOK
		} else {
			status = http.StatusInternalServerError
			w.WriteHeader(status)
		}

		// access log
		log.Trace("%s \"%s %s %s\" %d %s",
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
			r.Proto,
			status,
			time.Since(t1))

		// if handler returns nil, that means it will control the output
		if ret != nil {
			// pretty write json result
			w.Header().Set("Content-Type", "application/json")

			pretty, jsonErr := json.MarshalIndent(ret, "", "    ")
			if jsonErr != nil {
				ret = map[string]interface{}{"error": jsonErr.Error()}
				pretty, _ = json.MarshalIndent(ret, "", "    ")
			}
			w.Write(pretty)
		}
	})
}

func (e *Engine) decodeHTTPParams(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return params, nil
}
