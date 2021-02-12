package main

import (
	"github.com/LeakIX/l9format"
	"github.com/joho/godotenv"
)

type DotEnvHttpPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.WebPluginInterface {
	return DotEnvHttpPlugin{}
}

func (DotEnvHttpPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

var getEnvRequest = l9format.WebPluginRequest{
		Method: "GET",
		Path: "/.env",
		Headers: map[string]string{},
		Body:[]byte(""),
}

func (DotEnvHttpPlugin) GetRequests() []l9format.WebPluginRequest {
	return []l9format.WebPluginRequest{getEnvRequest}
}

func (DotEnvHttpPlugin) GetName() string {
	return "DotEnvConfigPlugin"
}

func (DotEnvHttpPlugin) GetStage() string {
	return "open"
}
func (plugin DotEnvHttpPlugin) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) (leak l9format.L9LeakEvent, hasLeak bool) {
	if !getEnvRequest.Equal(request) || response.Response.StatusCode != 200 {
		return leak, false
	}
	if len(response.Body) > 0 && response.Body[0] == '<' {
		return leak, false
	}
	envConfig, err := godotenv.Unmarshal(string(response.Body))
	if err != nil {
		return leak, false
	}

	if len(envConfig) > 1 && len(envConfig) < 2048 {
		leak.Data = string(response.Body)
		return leak, true
	}
	return leak, false
}
