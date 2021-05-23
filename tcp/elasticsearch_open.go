package tcp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LeakIX/l9format"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
	"strconv"
	"strings"
)

type ElasticSearchOpenPlugin struct {
	l9format.ServicePluginBase
}


func (ElasticSearchOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (ElasticSearchOpenPlugin) GetProtocols() []string {
	return []string{"elasticsearch", "kibana"}
}

func (ElasticSearchOpenPlugin) GetName() string {
	return "ElasticSearchOpenPlugin"
}

func (ElasticSearchOpenPlugin) GetStage() string {
	return "open"
}

// Get info
func (plugin ElasticSearchOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	log.Printf("Discovering %s ...", event.Url())
	url := "/_nodes"
	method := "GET"
	// Requires deep-http from l9tcpid
	if event.Protocol == "kibana" {
		majorVersion := 0
		versionSplit := strings.Split(event.Service.Software.Version, ".")
		if len(versionSplit) > 1 {
			majorVersion, _ = strconv.Atoi(versionSplit[0])
		}
		method = "POST"
		url = "/api/console/proxy?path=" + url2.QueryEscape("/_nodes") + "&method=GET"
		if majorVersion != 0 && majorVersion  < 5{
			method = "GET"
			url = "/elasticsearch/_nodes"
		}
		event.Summary += "Through Kibana endpoint\n"
		event.Service.Software.Name = "Kibana"
	} else {
		event.Service.Software.Name = "Elasticsearch"
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", event.Url(), url), nil)
	req.Header["User-Agent"] = []string{"l9plugin-ElasticSearchOpenPlugin/v1.0.0 (+https://leakix.net/)"}
	req.Header["kbn-xsrf"] = []string{"true"}
	if len(event.Service.Software.Version) > 3 && event.Protocol == "kibana" {
		req.Header["kbn-version"] = []string{event.Service.Software.Version}
	}
	if err != nil {
		log.Println("can't create request:", err)
		return false
	}
	// use the http client to fetch the page
	resp, err := plugin.GetHttpClient(ctx, event.Ip, event.Port).Do(req)
	if err != nil {
		log.Println("can't GET page:", err)
		return false
	}
	defer resp.Body.Close()
	httpReply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("can't read body:", err)
		return false
	}
	esReply := ElasticSearchCatNodesResponse{}
	err = json.Unmarshal(httpReply, &esReply)
	if err != nil {
		log.Println("can't parse body:", err)
		return false
	}
	// check if we got stg in tagline
	if len(esReply.Nodes) < 1 {
		return false
	}
	hasLeak = true
	log.Printf("Found %d nodes on ES endpoint, using first only", len(esReply.Nodes))
	for _, node := range esReply.Nodes {
		if event.Protocol == "elasticsearch" {
			event.Service.Software.Version = node.Version
		}
		event.Service.Software.OperatingSystem = node.OperatingSystem.Name + " " + node.OperatingSystem.Version
		break
	}
	// There's no index summary we can find in our reply, dispatch to explore a F** it :)
	event.Service.Credentials = l9format.ServiceCredentials{NoAuth: true}
	event.Summary += "NoAuth\n"
	event.Summary += "Cluster info:\n"
	event.Summary += string(httpReply)
	return true
}

// First thing we tried, turns out node API has more info we like
type ElasticSearchGreetResponse struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	ClusterUuid string `json:"cluster_uuid"`
	TagLine     string `json:"tagline"`
	Version     struct {
		Number                           string `json:"number"`
		BuildFlavor                      string `json:"build_flavor"`
		BuildType                        string `json:"build_type"`
		BuildHash                        string `json:"build_hash"`
		BuildDate                        string `json:"build_date"`
		BuildSnapshot                    bool   `json:"build_snapshot"`
		LuceneVersion                    string `json:"lucene_version"`
		MinimumWireCompatibilityVersion  string `json:"minimum_wire_compatibility_version"`
		MinimumIndexCompatibilityVersion string `json:"minimum_index_compatibility_version"`
	} `json:"version"`
}

type ElasticSearchCatNodesResponse struct {
	NodesSumary struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_nodes"`
	Nodes map[string]struct {
		Version         string `json:"version"`
		OperatingSystem struct {
			Name    string `json:"pretty_name"`
			Version string `json:"version"`
		} `json:"os"`
	} `json:"nodes"`
}
