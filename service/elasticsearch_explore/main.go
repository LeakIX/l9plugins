package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LeakIX/l9format"
	"github.com/LeakIX/l9format/utils"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	url2 "net/url"
	"strconv"
	"strings"
)

type ElasticSearchExplorePlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.ServicePluginInterface {
	return ElasticSearchExplorePlugin{}
}

func (ElasticSearchExplorePlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (ElasticSearchExplorePlugin) GetProtocols() []string {
	return []string{"elasticsearch","kibana"}
}

func (ElasticSearchExplorePlugin) GetName() string {
	return "ElasticSearchExplorePlugin"
}

func (ElasticSearchExplorePlugin) GetStage() string {
	return "explore"
}
// Get info
func (plugin ElasticSearchExplorePlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) (leak l9format.L9LeakEvent, hasLeak bool) {
	log.Printf("Discovering http://%s ...", net.JoinHostPort(event.Ip,event.Port))
	url := "/_cat/indices?format=json&bytes=b"
	method := "GET"
	if event.Protocol == "kibana" {
		majorVersion := 0
		versionSplit := strings.Split(event.Service.Software.Version, ".")
		if len(versionSplit) > 1 {
			majorVersion, _ = strconv.Atoi(versionSplit[0])
		}
		method= "POST"
		url = "/api/console/proxy?path=" + url2.QueryEscape("/_cat/indices?format=json&bytes=b") + "&method=GET"
		if majorVersion != 0 && majorVersion  < 5{
			method = "GET"
			url = "/elasticsearch/_cat/indices?format=json&bytes=b"
		}
		leak.Data += "Through Kibana endpoint\n"
	}
	scheme := "http"
	if event.HasTransport("tls"){
		scheme = "https"
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s://%s%s", scheme, net.JoinHostPort(event.Ip,event.Port), url), nil)
	req.Header["User-Agent"] = []string{"l9plugin-ElasticSearchExplorePlugin/0.1.0 (+https://leakix.net/)"}
	req.Header["kbn-xsrf"] = []string{"true"}
	if len(event.Service.Software.Version) > 3 {
		req.Header["kbn-version"] = []string{event.Service.Software.Version}
	}
	if err != nil {
		log.Println("can't create request:", err)
		return leak, false
	}
	// use the http client to fetch the page
	resp, err := plugin.GetHttpClient(ctx, event.Ip, event.Port).Do(req)
	if err != nil {
		log.Println("can't GET page:", err)
		return leak, false
	}
	defer resp.Body.Close()
	httpReply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("can't read body:", err)
		return leak, false
	}
	esReply := ElasticSearchCatIndicesResponse{}
	err = json.Unmarshal(httpReply, &esReply)
	if err != nil {
		log.Println("can't parse body:", err)
		return leak, false
	}
	// check if we got indices
	if len(esReply) < 1 {
		return leak, false
	}
	log.Printf("Found %d indices on ES endpoint", len(esReply))

	for _, esIndex := range esReply {
		if indexSize, err := strconv.ParseInt(esIndex.IndexSize, 10, 64); err == nil {
			leak.Data += fmt.Sprintf("Found index %s with %s documents (%s)\n", esIndex.Name, esIndex.DocCount, utils.HumanByteCount(indexSize))
		} else {
			leak.Data += fmt.Sprintf("Found index %s with %s documents (%s)\n", esIndex.Name, esIndex.DocCount, esIndex.IndexSize)
		}
		leak.Dataset.Collections++
		if docCount, err := strconv.ParseInt(esIndex.DocCount, 10, 64); err == nil {
			leak.Dataset.Rows += docCount
		}
		if indexSize, err := strconv.ParseInt(esIndex.IndexSize, 10, 64); err == nil {
			leak.Dataset.Size += indexSize
		}
		if strings.Contains(esIndex.Name, "meow")  || strings.Contains(esIndex.Name, "hello") ||
			strings.Contains(esIndex.Name, "readme") || strings.Contains(esIndex.Name, "read_me") {
			leak.Dataset.Infected = true
		}
	}
	leak.Data = fmt.Sprintf("Indices: %d, document count: %d, size: %s\n",
		leak.Dataset.Collections, leak.Dataset.Rows, utils.HumanByteCount(leak.Dataset.Size)) +
		leak.Data
	// Short leak, a second one will contain indices stats & co
	return leak, true
}

type ElasticSearchCatIndicesResponse []struct {
	Health    string `json:"health"`
	Status    string `json:"status"`
	Name      string `json:"index"`
	DocCount  string `json:"docs.count"`
	IndexSize string `json:"pri.store.size"`
}
