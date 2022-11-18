package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors"
)

type MetadataProcessor struct {
	PortalTags          []string        `toml:"portal_tags"`
	Timeout          config.Duration `toml:"timeout"`
	Log              telegraf.Logger `toml:"-"`
}

type PortalMetadata struct {
	Id int
}

var m PortalMetadata

type Response struct {

	Id			  	string `json:"id"`
	Cycle         	string `json:"cycle"`
	DeviceConfig   	string `json:"device_config"`
	GrindCycle		string `json:"grind_cycle"`
	SteamCycle      string `json:"steam_cycle"`
	WasteType      	string `json:"waste_type"`
	Type        	string `json:"type"`
	StartTime       string `json:"start_time"`
	EndTime      	string `json:"end_time"`
	Completed     	string `json:"completed"`
	Successful    	string `json:"successful"`
}

var meta_resp Response

const metadata_url = "http://169.254.169.254/metrics/metadata"

const sampleConfig = `
  ## Available tags to attach to metrics:
  ## * id
  ## * cycle
  ## * device_config
  ## * grind_cycle
  ## * steam_cycle
  ## * waste_type
  ## * type
  ## * start_time
  ## * end_time
  ## * completed
  ## * successful
  portal_tags = [ "id", "grind_cycle", "steam_cycle" ]
`

const (
	DefaultTimeout             = 10 * time.Second
)

func (r *MetadataProcessor) SampleConfig() string {
	return sampleConfig
}

func (r *MetadataProcessor) Description() string {
	return "Attach Portal metadata to metrics"
}

func (r *MetadataProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// add tags
	for _, metric := range in {
		r.Log.Debug("length is ",len(r.PortalTags))
		r.Log.Debug(r.PortalTags)
		for _, tag := range r.PortalTags {
			r.Log.Debug("checking tag=",tag)
			if v := getTagFromMetadataResponse(meta_resp, tag); v != "" {
				r.Log.Debug("adding tag=",tag," value=",v)
				metric.AddTag(tag, v)
			}
		}
	}
	return in
}

func (r *MetadataProcessor) Init() error {
	r.Log.Debug("Initializing Portal Metadata Processor")
	if len(r.PortalTags) == 0 {
		return errors.New("no tags specified in configuration")
	}
	meta_resp = getMetadata()
	r.Log.Debug(PrettyPrint(meta_resp))
	return nil
}

func init() {
	processors.Add("metadata", func() telegraf.Processor {
		return &MetadataProcessor{}
	})
}

func getMetadata() Response {
    resp, err := http.Get(metadata_url)
    if err != nil {
        fmt.Println("No response from request")
    }
    defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // response body is []byte

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {   // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return result
}

func getTagFromMetadataResponse(o Response, tag string) string {
	switch tag {
	case "id":
		return o.Id
	case "cycle":
		return o.Cycle
	case "device_config":
		return o.DeviceConfig
	case "grind_cycle":
		return o.GrindCycle
	case "steam_cycle":
		return o.SteamCycle
	case "start_time":
		return o.StartTime
	case "end_time":
		return o.EndTime
	case "completed":
		return o.Completed
	case "successful":
		return o.Successful
	default:
		return ""
	}
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
