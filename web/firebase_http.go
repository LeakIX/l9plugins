package web

import (
	"encoding/json"
	"strings"

	"github.com/LeakIX/l9format"
)

type FirebaseHttpPlugin struct {
	l9format.ServicePluginBase
}

func (FirebaseHttpPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (FirebaseHttpPlugin) GetRequests() []l9format.WebPluginRequest {
	return []l9format.WebPluginRequest{{
		Method:  "GET",
		Path:    "/.json",
		Headers: map[string]string{},
		Body:    []byte(""),
	}}
}

func (FirebaseHttpPlugin) GetName() string {
	return "FirebaseHttpPlugin"
}

func (FirebaseHttpPlugin) GetStage() string {
	return "open"
}
func (plugin FirebaseHttpPlugin) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	if !request.EqualAny(plugin.GetRequests())|| response.Response.StatusCode != 200  || !strings.HasSuffix(event.Host, ".firebaseio.com") {
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
		event.Leak.Dataset.Size = int64(len(response.Body))
		event.Summary = string(response.Body)
		return true
	}
	return false
}
