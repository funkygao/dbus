package engine

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/funkygao/golib/dag"
	"github.com/funkygao/golib/version"
	log "github.com/funkygao/log4go"
)

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
	output["ver"] = version.Version
	output["started"] = Globals().StartedAt
	output["elapsed"] = time.Since(Globals().StartedAt).String()
	output["pid"] = e.pid
	output["hostname"] = e.hostname
	output["revision"] = version.Revision
	return output, nil
}
