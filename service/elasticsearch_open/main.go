package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LeakIX/l9format"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	url2 "net/url"
	"strings"
)

type ElasticSearchOpenPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.ServicePluginInterface {
	return ElasticSearchOpenPlugin{}
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
func (plugin ElasticSearchOpenPlugin) Run(ctx context.Context, event *l9format.L9Event) (leak l9format.L9LeakEvent, hasLeak bool) {
	log.Printf("Discovering http://%s ...", net.JoinHostPort(event.Ip, event.Port))
	isKibana := false
	url := "/_nodes"
	method := "GET"
	kibanaVersionHeader := "0.0.0"
	// We have this if rerouted from HTTP :/
	for headerName, kbnVersion := range event.Http.Headers {
		if headerName == "Kbn-Version" && len(kbnVersion) == 1 {
			url = "/elasticsearch/_nodes"
			kibanaVersionHeader = kbnVersion
			versionSplit := strings.Split(kibanaVersionHeader, ".")
			if len(versionSplit) > 1 && versionSplit[0] != "2" && versionSplit[0] != "3" && versionSplit[0] != "4" && versionSplit[0] != "5" {
				method = "POST"
				url = "/api/console/proxy?path=" + url2.QueryEscape("/_nodes") + "&method=GET"
			}
			isKibana = true
			break
		}
		if headerName == "Kbn-Name" {
			//assume >= 7.x
			isKibana = true
			method = "POST"
			url = "/api/console/proxy?path=" + url2.QueryEscape("/_nodes") + "&method=GET"
		}
	}
	scheme := "http"
	if event.HasTransport("tls") {
		scheme = "https"
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s://%s:%s%s", scheme, event.Ip, event.Port, url), nil)
	req.Header["User-Agent"] = []string{"l9plugin-ElasticSearchOpenPlugin/0.1.1 (+https://leakix.net/)"}
	req.Header["kbn-xsrf"] = []string{"true"}

	if err != nil {
		log.Println("can't create request:", err)
		return leak, hasLeak
	}
	// use the http client to fetch the page
	resp, err := plugin.GetHttpClient(ctx, event.Ip, event.Port).Do(req)
	if err != nil {
		log.Println("can't GET page:", err)
		return leak, hasLeak
	}
	defer resp.Body.Close()
	httpReply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("can't read body:", err)
		return leak, hasLeak
	}
	esReply := ElasticSearchCatNodesResponse{}
	err = json.Unmarshal(httpReply, &esReply)
	if err != nil {
		log.Println("can't parse body:", err)
		return leak, hasLeak
	}
	// check if we got stg in tagline
	if len(esReply.Nodes) < 1 {
		return leak, hasLeak
	}
	hasLeak = true
	log.Printf("Found %d nodes on ES endpoint, using first only", len(esReply.Nodes))
	for _, node := range esReply.Nodes {
		event.Service.Software.Version = node.Version
		event.Service.Software.OperatingSystem = node.OperatingSystem.Name + " " + node.OperatingSystem.Version
		break
	}
	// There's no index summary we can find in our reply, dispatch to explore a F** it :)
	event.Service.Credentials = l9format.ServiceCredentials{NoAuth: true}

	event.Service.Software.Name = "Elasticsearch"

	leak.Data += "NoAuth\n"
	if isKibana {
		leak.Data += "Through Kibana endpoint\n"
		event.Service.Software.Name = "Kibana"
		if kibanaVersionHeader != "0.0.0" {
			event.Service.Software.Version = kibanaVersionHeader
		}
	}
	leak.Data += "Cluster info:\n"
	leak.Data += string(httpReply)
	return leak, hasLeak
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
