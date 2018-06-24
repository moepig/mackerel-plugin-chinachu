package mpchinachu

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var statsServer *httptest.Server
var stub string

var jsonStr = `{
  "connectedCount": 1,
  "feature": {
	"previewer": true,
    "streamer": true,
    "filer": true,
	"configurator": true
  },
  "system": {
    "core": 4
  },
  "operator": {
    "alive": true,
    "pid": 1122
  },
  "wui": {
    "alive": false,
	"pid": null
  }
}`

func TestGraphDefinition(t *testing.T) {
	var chinachu ChinachuPlugin

	graphdef := chinachu.GraphDefinition()
	if len(graphdef) != 2 {
		t.Errorf("GetTempfilename: %d should be 2", len(graphdef))
	}
}

func TestMain(m *testing.M) {
	os.Exit(mainTest(m))
}

func mainTest(m *testing.M) int {
	flag.Parse()

	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, stub)
			}))
	statsServer = ts

	log.Println("Started a stats server")

	return m.Run()
}

func TestFetchMetrics(t *testing.T) {
	// response a valid stats json
	stub = jsonStr

	// get metrics
	p := ChinachuPlugin{
		Target: strings.Replace(statsServer.URL, "http://", "", 1),
		Prefix: "chianchu",
	}
	metrics, err := p.FetchMetrics()
	if err != nil {
		t.Errorf("Failed to FetchMetrics: %s", err)
		return
	}

	// check the metrics
	expected := map[string]float64{
		"ConnectedCount": 1,
		"Previewer":      1,
		"Streamer":       1,
		"Filer":          1,
		"Configurator":   1,
	}

	for k, v := range expected {
		value, ok := metrics[k]
		if !ok {
			t.Errorf("metric of %s cannot be fetched", k)
			continue
		}
		if v != value {
			t.Errorf("metric of %s should be %v, but %v", k, v, value)
		}
	}
}

func TestFetchMetricsFail(t *testing.T) {
	p := ChinachuPlugin{
		Target: strings.Replace(statsServer.URL, "http://", "", 1),
		Prefix: "redash",
	}

	// return error against an invalid stats json
	stub = "{feature: [],}"
	_, err := p.FetchMetrics()
	if err == nil {
		t.Errorf("FetchMetrics should return error: stub=%v", stub)
	}
}
