package web

import (
	"encoding/json"
	"fmt"
	"github.com/LeakIX/l9format"
)

type WpUserEnumHttp struct {
	l9format.ServicePluginBase
}

func (WpUserEnumHttp) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (WpUserEnumHttp) GetRequests() []l9format.WebPluginRequest {
	return []l9format.WebPluginRequest{{
		Method: "GET",
		Path: "/?rest_route=/wp/v2/users/",
		Headers: map[string]string{},
		Body:[]byte(""),
		Tags: []string{"wordpress"},
	}}
}

func (WpUserEnumHttp) GetName() string {
	return "WpUserEnumHttp"
}

func (WpUserEnumHttp) GetStage() string {
	return "open"
}
func (plugin WpUserEnumHttp) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	if !request.EqualAny(plugin.GetRequests()) || response.Response.StatusCode != 200 {
		return false
	}
	var reply []WpUserReply
	err := json.Unmarshal(response.Body, &reply)
	if err != nil {
		return false
	}
	if len(reply) < 1 {
		return false
	}
	if len(reply[0].Name) == 0 && len(reply[0].Slug) == 0 {
		return false
	}
	event.Summary = "Found Wordpress users (CVE-2017-5487):\n\n"
	for _, user := range reply {
		event.Summary += fmt.Sprintf("User #%d %s\nName: %s\nUrl: %s\n\n", user.Id, user.Slug, user.Name, user.Url)
	}
	return true
}

type WpUserReply struct {
	Id int64
	Name string
	Url string
	Slug string
}