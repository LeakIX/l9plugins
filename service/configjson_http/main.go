package main

import (
	"encoding/json"
	"github.com/LeakIX/l9format"
)

type ConfigJsonHttp struct {
	l9format.ServicePluginBase
}

func New() l9format.WebPluginInterface {
	return ConfigJsonHttp{}
}

func (ConfigJsonHttp) GetVersion() (int, int, int) {
	return 0, 0, 1
}

var getServerStatus = l9format.WebPluginRequest{
		Method: "GET",
		Path: "/config.json",
		Headers: map[string]string{},
		Body:[]byte(""),
}

func (ConfigJsonHttp) GetRequests() []l9format.WebPluginRequest {
	return []l9format.WebPluginRequest{getServerStatus}
}

func (ConfigJsonHttp) GetName() string {
	return "ConfigJsonHttp"
}

func (ConfigJsonHttp) GetStage() string {
	return "open"
}
func (plugin ConfigJsonHttp) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	if !getServerStatus.Equal(request) || response.Response.StatusCode != 200 {
		return false
	}
	var reply CodeJsonReply
	var fullReply interface{}
	// It's a trap :/
	reply.Code = -323211
	reply.Status = reply.Code
	err := json.Unmarshal(response.Body, &reply)
	err = json.Unmarshal(response.Body, &fullReply)
	if err != nil {
		return false
	}
	if reply.Code != -323211 || reply.Status != reply.Code {
		return false
	}
	event.Leak.Dataset.Size = int64(len(response.Body))
	response.Body, err = json.MarshalIndent(fullReply,"", "  ")
	if err != nil {
		return false
	}
	event.Summary = string(response.Body)
	return true
}

type CodeJsonReply struct {
	Code int `json:"code"`
	Status int `json:"status"`
}