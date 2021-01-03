package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/LeakIX/l9format"
	"github.com/LeakIX/l9format/utils"
	"log"
	"net"
	"net/http"
	"strings"
)

type CouchDbOpenPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.ServicePluginInterface {
	return CouchDbOpenPlugin{}
}

func (CouchDbOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (CouchDbOpenPlugin) GetProtocols() []string {
	return []string{"couchdb"}
}

func (CouchDbOpenPlugin) GetName() string {
	return "CouchDbOpenPlugin"
}

func (CouchDbOpenPlugin) GetStage() string {
	return "open"
}
func (plugin CouchDbOpenPlugin) TestOpen(ctx context.Context, event *l9format.L9Event) bool {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/_membership", plugin.GetAddress(event)), nil)
	req.Header["User-Agent"] = []string{"TBI-CouchDbOpenPlugin/0.1.0 (+https://leakix.net/)"}
	req.Header["Content-Type"] = []string{"application/json"}
	if err != nil {
		log.Println("can't create request:", err)
		return false
	}
	resp, err := plugin.GetHttpClient(ctx, event.Ip, event.Port).Do(req)
	if err != nil {
		log.Println("can't GET page:", err)
		return false
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func (plugin CouchDbOpenPlugin) GetAddress(event *l9format.L9Event) string {
	address := net.JoinHostPort(event.Ip, event.Port)
	scheme := "http"
	if event.Host != "" {
		address = net.JoinHostPort(event.Host, event.Port)
	}
	if event.HasTransport("tls") {
		scheme = "https"
	}
	return scheme + "://" + address
}

// Get info
func (plugin CouchDbOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) (leak l9format.L9LeakEvent, hasLeak bool) {
	log.Printf("Discovering %s ...", plugin.GetAddress(event))
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/_all_dbs", plugin.GetAddress(event)), nil)
	req.Header["User-Agent"] = []string{"TBI-CouchDbOpenPlugin/0.2.0 (+https://leakix.net/)"}
	req.Header["Content-Type"] = []string{"application/json"}
	if err != nil {
		log.Println("can't create request:", err)
		return leak, false
	}
	resp, err := plugin.GetHttpClient(ctx, event.Ip, event.Port).Do(req)
	if err != nil {
		log.Println("can't GET page:", err)
		return leak, false
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		jsonDecoder := json.NewDecoder(resp.Body)
		var dbList []string
		err = jsonDecoder.Decode(&dbList)
		if err != nil {
			log.Println("can't parse body:", err)
			return leak, false
		}
		if len(dbList) > 0 {
			return plugin.GetInfos(ctx, event, dbList)
		}
	}
	return leak, false
}

func (plugin CouchDbOpenPlugin) GetDatabaseInfo(ctx context.Context, event *l9format.L9Event, dbNames []string) (dbInfo []DatabaseInfo) {
	reqBody, _ := json.Marshal(struct {
		Keys []string `json:"keys"`
	}{
		Keys: dbNames,
	})
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/_dbs_info", plugin.GetAddress(event)), bytes.NewReader(reqBody))
	req.Header["User-Agent"] = []string{"TBI-CouchDbOpenPlugin/0.1.0 (+https://leaks.nobody.run/)"}
	req.Header["Content-Type"] = []string{"application/json"}
	resp, err := plugin.GetHttpClient(ctx, event.Ip, event.Port).Do(req)
	if err != nil {
		log.Println("can't GET page:", err)
		return nil
	}
	jsonDecocer := json.NewDecoder(resp.Body)
	err = jsonDecocer.Decode(&dbInfo)
	resp.Body.Close()
	if err != nil {
		log.Println("can't read body:", err)
		return nil
	}
	return dbInfo
}

func (plugin CouchDbOpenPlugin) GetInfos(ctx context.Context, event *l9format.L9Event, dbList []string) (leak l9format.L9LeakEvent, hasLeak bool) {
	isOpen := plugin.TestOpen(ctx, event)
	if isOpen {
		leak.Data += "Weak auth\n"
	} else {
		leak.Data += "Schema only\n"
	}
	for _, dbNames := range plugin.chunkBy(dbList, 100) {
		for _, dbInfo := range plugin.GetDatabaseInfo(ctx, event, dbNames) {
			if isOpen {
				leak.Dataset.Collections++
				leak.Dataset.Rows += dbInfo.Info.DocCount
				leak.Dataset.Size += dbInfo.Info.DiskSize
			}
			leak.Data += fmt.Sprintf("Found table %s with %d documents (%s)\n", dbInfo.Info.Name, dbInfo.Info.DocCount, utils.HumanByteCount(dbInfo.Info.DiskSize))
			if strings.Contains(dbInfo.Info.Name, "read_me") || strings.Contains(dbInfo.Info.Name, "wegeturdb") || strings.Contains(dbInfo.Info.Name, "meow") {
				leak.Dataset.Infected = true
			}
		}
	}
	leak.Data = fmt.Sprintf("Databases: %d, document count: %d, size: %s\n",
		leak.Dataset.Collections, leak.Dataset.Rows, utils.HumanByteCount(leak.Dataset.Size)) +
		leak.Data
	return leak, true
}

type DatabaseInfo struct {
	Info struct {
		Name     string `json:"db_name"`
		DocCount int64  `json:"doc_count"`
		DiskSize int64  `json:"disk_size"`
	} `json:"info"`
}

func (plugin CouchDbOpenPlugin) chunkBy(items []string, chunkSize int) (chunks [][]string) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}
