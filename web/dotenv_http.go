package web

import (
	"github.com/LeakIX/l9format"
	"github.com/joho/godotenv"
	"regexp"
	"strings"
)

type DotEnvHttpPlugin struct {
	l9format.ServicePluginBase
}

func (DotEnvHttpPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (DotEnvHttpPlugin) GetRequests() []l9format.WebPluginRequest {
	return []l9format.WebPluginRequest{{
		Method:  "GET",
		Path:    "/.env",
		Headers: map[string]string{},
		Body:    []byte(""),
	}}
}

func (DotEnvHttpPlugin) GetName() string {
	return "DotEnvConfigPlugin"
}

func (DotEnvHttpPlugin) GetStage() string {
	return "open"
}
func (plugin DotEnvHttpPlugin) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	if !request.EqualAny(plugin.GetRequests()) || response.Response.StatusCode != 200 {
		return false
	}
	if len(response.Body) > 0 && response.Body[0] == '<' {
		return false
	}
	envConfig, err := godotenv.Unmarshal(string(response.Body))
	if err != nil {
		return false
	}
	if len(envConfig) > 1 && len(envConfig) < 2048 {
		event.Summary = string(response.Body)
		event.Leak.Severity = l9format.SEVERITY_MEDIUM
		if len(checkSensitiveKeyPatterns(envConfig)) > 0 {
			event.Leak.Severity = l9format.SEVERITY_HIGH
		}
		if awsFakeKey, hasKey := envConfig["AWS_ACCESS_KEY_ID"]; hasKey && len(envConfig) == 3 && strings.HasPrefix(awsFakeKey, "ASIAXM") {
			event.Leak.Severity = l9format.SEVERITY_LOW
		}
		return true
	}
	return false
}

var sensitiveKeyPatterns = []string{
	"^aws_.*",
	".*_password$",
	".*_key$",
	"^mysql_.*",
	".*secret.*",
	".*private.*",
}

func checkSensitiveKeyPatterns(config map[string]string) (matches []string) {
	for configKey, configValue := range config {
		for _, sensitiveKeyPattern := range sensitiveKeyPatterns {
			if match, err := regexp.MatchString(sensitiveKeyPattern, strings.ToLower(configKey)); err == nil && match && len(configValue) > 0 {
				matches = append(matches, configKey)
				// Don't match 2 patterns, proceed to next file
				break
			}
		}
	}
	return matches
}
