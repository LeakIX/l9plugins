package web

/*
 * Thanks to https://twitter.com/HaboubiAnis for pointers on this plugin.
 */

import (
	"fmt"
	"github.com/LeakIX/l9format"
	"github.com/PuerkitoBio/goquery"
	"github.com/hashicorp/go-version"
	"log"
)

type ConfluenceVersionIssue struct {
	l9format.ServicePluginBase
}

func (ConfluenceVersionIssue) GetVersion() (int, int, int) {
	return 0, 1, 0
}

func (ConfluenceVersionIssue) GetName() string {
	return "ConfluenceVersionIssue"
}

func (ConfluenceVersionIssue) GetStage() string {
	return "open"
}

var ConfluenceRequest = l9format.WebPluginRequest{Method: "GET", Path: "/login.action"}

func (plugin ConfluenceVersionIssue) GetRequests() []l9format.WebPluginRequest {
	//No need to register requests, / is a default one
	return []l9format.WebPluginRequest{
		ConfluenceRequest,
	}
}

// Test leak from response
func (plugin ConfluenceVersionIssue) Verify(request l9format.WebPluginRequest, response l9format.WebPluginResponse, event *l9format.L9Event, options map[string]string) bool {
	// If we're checking a /login.action request, negative
	if !ConfluenceRequest.Equal(request) {
		log.Printf("skipping confluence %s", request.Path)
		return false
	}
	// If there's no HTML in the response, negative
	if response.Document == nil {
		log.Println("skipping confluence")
		//no html
		return false
	}
	// Find metas
	response.Document.Find("meta").Each(func(i int, selection *goquery.Selection) {
		// With ajs-version-number name attrib
		if attrName, hasAttr := selection.Attr("name"); hasAttr && attrName == "ajs-version-number" {
			if attrValue, hasAttrValue := selection.Attr("content"); hasAttrValue {
				log.Printf("Found version %s", attrValue)
				confluenceVersion, err := version.NewVersion(attrValue)
				if err != nil {
					return
				}
				if len(confluenceVersion.Segments()) < 3 {
					return
				}
				event.Service.Software.Name = "Confluence"
				event.Service.Software.Version = confluenceVersion.String()
				segments := confluenceVersion.Segments()
				major, minor, patch := segments[0], segments[1], segments[2]
				// version test, TODO didn't care optimizing the logic, make it better for next issues
				if major < 6 && major > 4 {
					event.Summary = fmt.Sprintf("Confluence version %s likely vulnerable to CVE-2021-26084", confluenceVersion.String())
					event.AddTag("cve-2021-26084")
					return
				}
				if major == 6 && minor < 13 {
					event.Summary = fmt.Sprintf("Confluence version %s likely vulnerable to CVE-2021-26084", confluenceVersion.String())
					event.AddTag("cve-2021-26084")
					return
				}
				if major == 6 && minor == 13 {
					if patch < 23 {
						event.Summary = fmt.Sprintf("Confluence version %s likely vulnerable to CVE-2021-26084", confluenceVersion.String())
						event.AddTag("cve-2021-26084")
					}
					return
				}
				if major == 6 && minor > 13 {
					event.Summary = fmt.Sprintf("Confluence version %s likely vulnerable to CVE-2021-26084", confluenceVersion.String())
					event.AddTag("cve-2021-26084")
					return
				}
				if major == 7 && minor == 4 {
					if patch < 11 {
						event.Summary = fmt.Sprintf("Confluence version %s likely vulnerable to CVE-2021-26084", confluenceVersion.String())
						event.AddTag("cve-2021-26084")
					}
					return
				}
				if major == 7 && minor > 4 && minor < 11 {
					event.Summary = fmt.Sprintf("Confluence version %s likely vulnerable to CVE-2021-26084", confluenceVersion.String())
					event.AddTag("cve-2021-26084")
					return
				}
				if major == 7 && minor == 11 {
					if patch < 6 {
						event.Summary = fmt.Sprintf("Confluence version %s likely vulnerable to CVE-2021-26084", confluenceVersion.String())
						event.AddTag("cve-2021-26084")
					}
					return
				}
				if major == 7 && minor == 12 {
					if patch < 5 {
						event.Summary = fmt.Sprintf("Confluence version %s likely vulnerable to CVE-2021-26084", confluenceVersion.String())
						event.AddTag("cve-2021-26084")
					}
					return
				}
			}
		}
	})
	return event.HasTag("cve-2021-26084")
}
