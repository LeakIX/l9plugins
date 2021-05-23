package main

import (
	"encoding/json"

	"github.com/LeakIX/l9format"
)

type FirebaseHttpPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.WebPluginInterface {
	return FirebaseHttpPlugin{}
}

func (FirebaseHttpPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

var getEnvRequest = l9format.WebPluginRequest{
	Method:  "GET",
	Path:    "/.json",
	Headers: map[string]string{},
	Body:    []byte(""),
}

func (FirebaseHttpPlugin) GetRequests() []l9format.WebPluginRequest {
	return []l9format.WebPluginRequest{getEnvRequest}
}

func (FirebaseHttpPlugin) GetName() string {
	return "DotEnvConfigPlugin"
}

func (FirebaseHttpPlugin) GetStage() string {
	return "open"
}
func (plugin FirebaseHttpPlugin) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	if !getEnvRequest.Equal(request) || response.Response.StatusCode != 200 {
		return false
	}
	if len(response.Body) > 0 && response.Body[0] == '<' {
		return false
	}
	var dat map[string]interface{}
	err := json.Unmarshal(response.Body, &dat)

	if err != nil {
		return false
	}

	if len(dat) > 0 {
		return true
	}

	return false
}
