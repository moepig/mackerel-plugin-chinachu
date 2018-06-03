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
	Target   string
	Tempfile string
}

type Status struct {
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
		Label: "Connected",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "connectedcount", Label: "Count", Type: "uint32"},
		},
	},
	"chinachu.feature": mp.Graphs{
		Label: "Feature",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "previewer", Label: "Previewer", Type: "uint32"},
			{Name: "streamer", Label: "Streamer", Type: "uint32"},
			{Name: "filer", Label: "Filer", Type: "uint32"},
			{Name: "configurator", Label: "Configurator", Type: "uint32"},
		},
	},
}

// FetchMetrics interface for mackerelplugin
func (m ChinachuPlugin) FetchMetrics() (map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s/api/status.json", m.Target)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	byteArray, _ := ioutil.ReadAll(response.Body)

	var status Status
	if err := json.Unmarshal(byteArray, &status); err != nil {
		log.Fatal(err)
	}

	stat := make(map[string]interface{})

	stat["connected_count"] = status.ConnectedCount

	stat["previewer"] = Bool2Int(status.Feature.Previewer)
	stat["streamer"] = Bool2Int(status.Feature.Streamer)
	stat["filer"] = Bool2Int(status.Feature.Filer)
	stat["configurator"] = Bool2Int(status.Feature.Configurator)

	return stat, err
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

func Do() {
	optHost := flag.String("host", "", "chinachu-wui hostname")
	optPort := flag.String("port", "", "chinachu-wui port")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	var plugin ChinachuPlugin

	plugin.Target = fmt.Sprintf("%s:%s", *optHost, *optPort)

	helper := mp.NewMackerelPlugin(plugin)

	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.Tempfile = "/tmp/.mackerel-plugin-chinachu"
	}

	helper.Run()
}
