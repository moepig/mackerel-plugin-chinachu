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

	"github.com/gorilla/mux"
)

var statusServer *httptest.Server
var stub map[string]string

var jsonStr = map[string]string{
	"status": `{
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
}`,
	"recorded": `[
  {
	"id": "36tfa8hbsd",
	"category": "anime",
	"title": "test title",
	"fullTitle": "test title #1 「ほげ」",
	"detail": "detail",
	"start":1507390200000, 
	"end":1507392000000,
	"seconds":1800, 
	"description": "detail",
	"extra": {
		"番組内容": "ばんぐみないよう",
		"あらすじ◇": "あらすじ",
		"出演者": "うんこ太郎"
	},
	"channel": {
	  "type": "GR",
	  "channel": "16",
	  "name": "ＴＯＫＹＯ　ＭＸ１",
	  "id": "1hkhnrs",
	  "sid": 23608,
	  "nid": 32391,
	  "hasLogoData": true,
	  "n": 32
	},
	"subTitle": "ほげ",
	"episode": 1,
	"flags": [""],
	"isConflict": false,
	"recordedFormat": "",
	"priority":2, 
	"tuner": {
	  "name":"Mirakurun (UnixSocket)",
	  "command":"*",
	  "isScrambling":false
	},
	"command":"mirakurun type=GR url=/api/programs/323912360850925/stream?decode=1 priority=2",
	"recorded":"/opt/tv/1.m2ts"
  },
  {
	"id": "36tfa8hbse",
	"category": "anime",
	"title": "test title",
	"fullTitle": "test title #2 「ふが」",
	"detail": "detail",
	"start":1507390200001, 
	"end":1507392000001,
	"seconds":1800, 
	"description": "detail",
	"extra": {
		"番組内容": "ばんぐみないよう",
		"あらすじ◇": "あらすじ",
		"出演者": "うんこ太郎"
	},
	"channel": {
	  "type": "GR",
	  "channel": "16",
	  "name": "ＴＯＫＹＯ　ＭＸ１",
	  "id": "1hkhnrs",
	  "sid": 23608,
	  "nid": 32391,
	  "hasLogoData": true,
	  "n": 32
	},
	"subTitle": "ふが",
	"episode": 2,
	"flags": [""],
	"isConflict": false,
	"recordedFormat": "",
	"priority":2,
	"tuner": {
	  "name":"Mirakurun (UnixSocket)",
	  "command":"*",
	  "isScrambling":false
	},
	"command":"mirakurun type=GR url=/api/programs/323912360850925/stream?decode=1 priority=2",
	"recorded":"/opt/tv/2.m2ts"
  }
]`,
}

func TestGraphDefinition(t *testing.T) {
	var chinachu ChinachuPlugin

	graphdef := chinachu.GraphDefinition()
	if len(graphdef) != 3 {
		t.Errorf("GetTempfilename: %d should be 2", len(graphdef))
	}
}

func TestMain(m *testing.M) {
	os.Exit(mainTest(m))
}

func mainTest(m *testing.M) int {
	flag.Parse()

	router := mux.NewRouter()
	router.HandleFunc(
		"/api/status.json", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, stub["status"])
			}))
	router.HandleFunc(
		"/api/recorded.json", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, stub["recorded"])
			}))
	statusServer = httptest.NewServer(router)
	log.Println("Started a stats server")

	return m.Run()
}

func TestFetchMetrics(t *testing.T) {
	// response a valid stats json
	stub = jsonStr

	// get metrics
	p := ChinachuPlugin{
		Target: strings.Replace(statusServer.URL, "http://", "", 1),
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
		"RecordedCount":  2,
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
		Target: strings.Replace(statusServer.URL, "http://", "", 1),
		Prefix: "redash",
	}

	// return error against an invalid stats json
	stub = map[string]string{
		"status":   "{feature: [],}",
		"recorded": "[]",
	}
	_, err := p.FetchMetrics()
	if err == nil {
		t.Errorf("FetchMetrics should return error: stub=%v", stub)
	}
}
