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
	ConnectedCount int `json:"connectedCount"`
}

var graphdef = map[string]mp.Graphs{
	"chinachu.recording_count": mp.Graphs{
		Label: "Recording Count",
		Unit:  "integer",
		Metrics: [](mp.Metrics){
			mp.Metrics{Name: "RecordingCount", Label: "RecordingCount", Type: "integer"},
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
	stat["recording_count"] = status.ConnectedCount

	return stat, err
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
