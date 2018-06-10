package mpchinachu

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	"github.com/mackerelio/golib/logging"
)

var logger = logging.GetLogger("metrics.plugin.chinachu")

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

var graphdef = map[string]mp.Graphs{
	"chinachu.connected_count": mp.Graphs{
		Label: "Connected Count",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "ConnectedCount", Label: "Count", Diff: false, Type: "uint32"},
		},
	},
	"chinachu.feature": mp.Graphs{
		Label: "Feature",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "Previewer", Label: "Previewer", Diff: false, Type: "uint32"},
			{Name: "Streamer", Label: "Streamer", Diff: false, Type: "uint32"},
			{Name: "Filer", Label: "Filer", Diff: false, Type: "uint32"},
			{Name: "Configurator", Label: "Configurator", Diff: false, Type: "uint32"},
		},
	},
}

// GetStatus サーバー情報取得
// https://github.com/Chinachu/Chinachu/wiki/REST-API#-status
func GetStatus(host string) (status, error) {
	url := fmt.Sprintf("http://%s/api/status.json", host)

	var s status
	response, err := http.Get(url)

	if err != nil {
		return s, err
	}
	defer response.Body.Close()

	byteArray, _ := ioutil.ReadAll(response.Body)

	if err := json.Unmarshal(byteArray, &s); err != nil {
		log.Fatal(err)
	}

	return s, err
}

// FetchMetrics interface for mackerelplugin
func (m ChinachuPlugin) FetchMetrics() (map[string]interface{}, error) {
	stat := make(map[string]interface{})

	status, err := GetStatus(m.Target)
	if err != nil {
		return nil, err
	}

	stat["ConnectedCount"] = status.ConnectedCount

	stat["Previewer"] = Bool2Int(status.Feature.Previewer)
	stat["Streamer"] = Bool2Int(status.Feature.Streamer)
	stat["Filer"] = Bool2Int(status.Feature.Filer)
	stat["Configurator"] = Bool2Int(status.Feature.Configurator)

	return stat, nil
}

func Bool2Int(x bool) int {
	if x {
		return 1
	}
	return 0
}

// GraphDefinition interface for mackerelplugin
func (m ChinachuPlugin) GraphDefinition() map[string](mp.Graphs) {
	return graphdef
}

// MetricKeyPrefix interface for mackerelplugin
func (m ChinachuPlugin) MetricKeyPredix() string {
	if m.Prefix == "" {
		m.Prefix = "chinachu"
	}
	return m.Prefix
}

func Do() {
	optPrefix := flag.String("matric-key-prefix", "chinachu", "Metric key prefix")
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
