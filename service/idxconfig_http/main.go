package main

import (
	"github.com/LeakIX/IndexOfBrowser"
	"github.com/LeakIX/l9format"
	"log"
	"strings"
)

type IdxConfigPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.WebPluginInterface {
	return IdxConfigPlugin{}
}

func (IdxConfigPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

var idxConfigRequest = l9format.WebPluginRequest{
		Method: "GET",
		Path: "/idx_config/",
		Headers: map[string]string{},
		Body:[]byte(""),
}

func (IdxConfigPlugin) GetRequests() []l9format.WebPluginRequest {
	return []l9format.WebPluginRequest{idxConfigRequest}
}

func (IdxConfigPlugin) GetName() string {
	return "IdxConfigPlugin"
}

func (IdxConfigPlugin) GetStage() string {
	return "open"
}
func (plugin IdxConfigPlugin) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) ( hasLeak bool) {
	if !idxConfigRequest.Equal(request) || response.Response.StatusCode != 200 || response.Document == nil {
		return false
	}
	if strings.HasPrefix(response.Document.Find("title").Text(), "Index of /") {
		browser := IndexOfBrowser.NewBrowser(event.Url())
		log.Println(event.Url())
		files, err := browser.Ls()
		if err != nil {
			return false
		}
		for _, file := range files {
			event.Summary += "Found " + file.Name + "\n"
			event.Leak.Dataset.Files++
			event.Leak.Dataset.Infected = true
		}
		if len(files) > 0 {
			return  true
		}
	}
	return false
}
