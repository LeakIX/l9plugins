package main

import (
	"github.com/LeakIX/IndexOfBrowser"
	"github.com/LeakIX/l9format"
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
func (plugin IdxConfigPlugin) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) (leak l9format.L9LeakEvent, hasLeak bool) {
	if !idxConfigRequest.Equal(request) || response.Response.StatusCode != 200 || response.Document == nil {
		return leak, false
	}
	if strings.HasPrefix("Index of /", response.Document.Find("title").Text()) {
		browser := IndexOfBrowser.NewBrowser(response.Response.Request.URL.String())
		files, err := browser.Ls()
		if err != nil {
			return leak, false
		}
		for _, file := range files {
			leak.Data += "Found " + file.Name + "\n"
			leak.Dataset.Files++
			leak.Dataset.Infected = true
		}
		if len(files) > 0 {
			return leak, true
		}
	}
	return leak, false
}
