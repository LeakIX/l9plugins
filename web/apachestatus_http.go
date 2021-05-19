package web

import (
	"github.com/LeakIX/l9format"
)

type ApacheStatusHttpPlugin struct {
	l9format.ServicePluginBase
}

func (ApacheStatusHttpPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

var getServerStatus = l9format.WebPluginRequest{
		Method: "GET",
		Path: "/server-status",
		Headers: map[string]string{},
		Body:[]byte(""),
}

func (ApacheStatusHttpPlugin) GetRequests() []l9format.WebPluginRequest {
	return []l9format.WebPluginRequest{getServerStatus}
}

func (ApacheStatusHttpPlugin) GetName() string {
	return "ApacheStatusHttpPlugin"
}

func (ApacheStatusHttpPlugin) GetStage() string {
	return "open"
}
func (plugin ApacheStatusHttpPlugin) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	if !getServerStatus.Equal(request) || response.Response.StatusCode != 200  || response.Document == nil {
		return false
	}
	if response.Document.Find("title").Text() == "Apache Status" {
		event.Summary = response.Document.Text()
		return true
	}
	return false
}