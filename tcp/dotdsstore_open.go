package tcp

import (
	"context"
	"fmt"
	"github.com/LeakIX/l9format"
	"github.com/gehaxelt/ds_store"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type DotDsStoreOpenPlugin struct {
	l9format.ServicePluginBase
}

func (DotDsStoreOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (DotDsStoreOpenPlugin) GetProtocols() []string {
	return []string{"http", "https"}
}

func (DotDsStoreOpenPlugin) GetName() string {
	return "DotDsStoreOpenPlugin"
}

func (DotDsStoreOpenPlugin) GetStage() string {
	return "open"
}

func (plugin DotDsStoreOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	log.Printf("Discovering %s ...", event.Url())
	results := plugin.getDsStoreFiles(ctx, event, event.Url(), "/")
	if len(results) > 0 {
		event.Summary = fmt.Sprintf("Found %d files trough .DS_Store spidering:\n\n", len(results))
		event.Summary += strings.Join(results, "\n")
		event.Leak.Severity = l9format.SEVERITY_LOW
		if len(results) > 32 {
			event.Leak.Severity = l9format.SEVERITY_MEDIUM
		}
		if len(checkSensitiveFilePatterns(results)) > 0 {
			event.Leak.Severity = l9format.SEVERITY_HIGH
		}
		return true
	}
	return false
}

var sensitiveFilePatterns = []string{
	".*back$",
	".*backup$",
	".*swp$",
	".*sql$",
	".*zip$",
	".*gz$",
	".*rar$",
	".*7z$",
	".*_history$",
}

func checkSensitiveFilePatterns(files []string) (matches []string) {
	for _, file := range files {
		for _, sensitiveFilePattern := range sensitiveFilePatterns {
			if match, err := regexp.MatchString(sensitiveFilePattern, file); err == nil && match {
				matches = append(matches, file)
				// Don't match 2 patterns, proceed to next file
				break
			}
		}
	}
	return matches
}

func (plugin DotDsStoreOpenPlugin) getDsStoreFiles(ctx context.Context, event *l9format.L9Event, rootUrl, path string) (results []string) {
	history := make(map[string]bool)
	if strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
		path = strings.Trim(path, "/")
	}
	if !strings.HasSuffix(rootUrl, "/") && len(path) > 0 {
		rootUrl += "/"
	}
	checkUrl := rootUrl + path + "/.DS_Store"
	log.Printf("Checking %s", checkUrl)
	req, err := http.NewRequest("GET", checkUrl, nil)
	if err != nil {
		log.Println(err)
		return results
	}
	resp, err := plugin.GetHttpClient(ctx, event.Ip, event.Port).Do(req)
	if err != nil {
		log.Println(err)
		return results
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		log.Println(err)
		return
	}
	allocator, err := ds_store.NewAllocator(bodyBytes)
	if err != nil {
		log.Println(err)
		return
	}
	filenames, err := allocator.TraverseFromRootNode()
	if err != nil {
		log.Println(err)
		return
	}
	for _, filename := range filenames {
		if filename == "." {
			continue
		}
		if _, found := history[filename]; !found {
			history[filename] = true
		} else {
			continue
		}
		event.Leak.Dataset.Files++
		if event.Leak.Dataset.Files > 128 {
			return results
		}
		if len(path) > 0 {
			results = append(results, "/"+path+"/"+filename)
		} else {
			results = append(results, "/"+filename)
		}
		if !strings.Contains(filename, ".") || strings.HasPrefix(filename, ".") {
			results = append(results, plugin.getDsStoreFiles(ctx, event, rootUrl, path+"/"+filename)...)
		}
	}
	return results
}
