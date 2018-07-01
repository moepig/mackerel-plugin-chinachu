package mpchinachu

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

type ChinachuPlugin struct {
	Prefix   string
	Target   string
	Tempfile string
}

type status struct {
	ConnectedCount int     `json:"connectedCount"`
	Feature        feature `json:"feature"`
}

type feature struct {
	Previewer    bool
	Streamer     bool
	Filer        bool
	Configurator bool
}

type recorded struct {
	ID string `json:"id"`
}

var graphdef = map[string]mp.Graphs{
	"connected_count": mp.Graphs{
		Label: "Chinachu - Connected Count",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "ConnectedCount", Label: "Count", Diff: false},
		},
	},
	"feature": mp.Graphs{
		Label: "Chinachu - Feature",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "Previewer", Label: "Previewer", Diff: false},
			{Name: "Streamer", Label: "Streamer", Diff: false},
			{Name: "Filer", Label: "Filer", Diff: false},
			{Name: "Configurator", Label: "Configurator", Diff: false},
		},
	},
	"recorded": mp.Graphs{
		Label: "Chinachu - Recorded",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "RecordedCount", Label: "RecordedCount", Diff: false},
		},
	},
}

// RequestAPI サーバー情報取得
// https://github.com/Chinachu/Chinachu/wiki/REST-API#-status
func requestAPI(host string, path string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/api/%s.json", host, path)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	byteArray, err := ioutil.ReadAll(response.Body)

	return byteArray, err
}

func GetStatus(host string) (status, error) {
	var s status
	byteArray, err := requestAPI(host, "status")

	err = json.Unmarshal(byteArray, &s)
	return s, err
}

func GetRecorded(host string) ([]recorded, error) {
	var r []recorded
	byteArray, err := requestAPI(host, "recorded")

	err = json.Unmarshal(byteArray, &r)
	return r, err
}

// FetchMetrics interface for mackerelplugin
func (m ChinachuPlugin) FetchMetrics() (map[string]float64, error) {
	stat := make(map[string]float64)

	status, err := GetStatus(m.Target)
	if err != nil {
		return nil, err
	}

	recorded, err := GetRecorded(m.Target)
	if err != nil {
		return nil, err
	}

	stat["ConnectedCount"] = float64(status.ConnectedCount)

	stat["Previewer"] = float64(Bool2Int(status.Feature.Previewer))
	stat["Streamer"] = float64(Bool2Int(status.Feature.Streamer))
	stat["Filer"] = float64(Bool2Int(status.Feature.Filer))
	stat["Configurator"] = float64(Bool2Int(status.Feature.Configurator))

	stat["RecordedCount"] = float64(len(recorded))

	return stat, nil
}

// Bool2Int bool -> int 1 or 0
func Bool2Int(x bool) int {
	if x {
		return 1
	}
	return 0
}

// GraphDefinition interface for mackerelplugin
func (m ChinachuPlugin) GraphDefinition() map[string]mp.Graphs {
	return graphdef
}

// MetricKeyPrefix interface for mackerelplugin
func (m ChinachuPlugin) MetricKeyPrefix() string {
	if m.Prefix == "" {
		m.Prefix = "chinachu"
	}
	return m.Prefix
}

func Do() {
	optPrefix := flag.String("metric-key-prefix", "chinachu", "Metric key prefix")
	optHost := flag.String("host", "", "chinachu-wui hostname")
	optPort := flag.String("port", "", "chinachu-wui port")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	var plugin ChinachuPlugin

	plugin.Target = fmt.Sprintf("%s:%s", *optHost, *optPort)
	plugin.Prefix = *optPrefix

	helper := mp.NewMackerelPlugin(plugin)

	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.Tempfile = "/tmp/.mackerel-plugin-chinachu"
	}

	helper.Run()
}
