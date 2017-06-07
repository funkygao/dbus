package engine

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

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
		e.apiListener = nil

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
	e.RegisterAPI("/api/v1/queues", e.handleQueuesV1).Methods("GET")
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
		} else if err == ErrInvalidParam {
			status = http.StatusBadRequest
			w.WriteHeader(status)
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
