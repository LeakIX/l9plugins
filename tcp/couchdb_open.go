package tcp

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

// Gets a database list and runs futher steps
func (plugin CouchDbOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	log.Printf("Discovering %s ...", plugin.GetAddress(event))
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/_all_dbs", plugin.GetAddress(event)), nil)
	req.Header["User-Agent"] = []string{"TBI-CouchDbOpenPlugin/0.2.0 (+https://leakix.net/)"}
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
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		jsonDecoder := json.NewDecoder(resp.Body)
		var dbList []string
		err = jsonDecoder.Decode(&dbList)
		if err != nil {
			log.Println("can't parse body:", err)
			return false
		}
		if len(dbList) > 0 {
			return plugin.GetInfos(ctx, event, dbList)
		}
	}
	return false
}

// Iterate over the database list to get more informations
func (plugin CouchDbOpenPlugin) GetInfos(ctx context.Context, event *l9format.L9Event, dbList []string) (hasLeak bool) {
	isOpen := plugin.TestOpen(ctx, event)
	if isOpen {
		event.Summary += "Weak auth\n"
		event.Leak.Severity = l9format.SEVERITY_MEDIUM
	} else {
		event.Summary += "Schema only\n"
		event.Leak.Severity = l9format.SEVERITY_LOW
	}
	for _, dbNames := range plugin.chunkBy(dbList, 50) {
		for _, dbInfo := range plugin.GetDatabaseInfo(ctx, event, dbNames) {
			if isOpen {
				event.Leak.Dataset.Collections++
				event.Leak.Dataset.Rows += dbInfo.Info.DocCount
				event.Leak.Dataset.Size += dbInfo.Info.DiskSize
			}
			event.Summary += fmt.Sprintf("Found table %s with %d documents (%s)\n", dbInfo.Info.Name, dbInfo.Info.DocCount, utils.HumanByteCount(dbInfo.Info.DiskSize))
			if (strings.HasPrefix(dbInfo.Info.Name, "read") && strings.HasSuffix(dbInfo.Info.Name, "me")) || strings.Contains(dbInfo.Info.Name, "wegeturdb") || strings.Contains(dbInfo.Info.Name, "meow") {
				event.Leak.Dataset.Infected = true
			}
		}
	}
	if event.Leak.Dataset.Rows > 1000 {
		event.Leak.Severity = l9format.SEVERITY_HIGH
		if event.Leak.Dataset.Infected {
			event.Leak.Severity = l9format.SEVERITY_CRITICAL
		}
	}
	if event.Leak.Dataset.Rows > 100000 {
		event.Leak.Severity = l9format.SEVERITY_CRITICAL
	}
	event.Summary = fmt.Sprintf("Databases: %d, document count: %d, size: %s\n",
		event.Leak.Dataset.Collections, event.Leak.Dataset.Rows, utils.HumanByteCount(event.Leak.Dataset.Size)) +
		event.Summary
	return true
}

// Check if data accessible or if only the schema is exposed
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

// Get database information from a list of names
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

// Helper to generate HTTP address
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

// Minimal structure returned from the info endpoint
type DatabaseInfo struct {
	Info struct {
		Name     string `json:"db_name"`
		DocCount int64  `json:"doc_count"`
		DiskSize int64  `json:"disk_size"`
	} `json:"info"`
}

// Splits a list by chunks
func (plugin CouchDbOpenPlugin) chunkBy(items []string, chunkSize int) (chunks [][]string) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}
