// Eventually I would like to create structs for each type (json, xml) and give
// them methods to crawl the urls they contain. Then I could reduce some of the
// duplicate code.
package checker

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"

	"github.com/bythepixel/urlchecker/pkg/client"
)

type Messager interface {
	SendMessage(status int, url string, message string)
}

type HealthCheck struct {
	// Path to check
	Path string `json:"path"`

	// Regex used to check the body of the response
	Regex string `json:"regex"`

	// Status code expected from URL
	Status int `json:"status"`

	// XMLSitemap indicates the Path to check is an XML Sitemap
	XMLSitemap bool `json:"xml_sitemap"`
}

type XMLSitemap struct {
	URL []struct {
		Location string `xml:"loc"`
	} `xml:"url"`
}

func Check(filename, protocol, hostname string, messager Messager) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err.Error())
	}

	var urls []HealthCheck
	err = json.Unmarshal(bytes, &urls)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, check := range urls {
		url := protocol + "://" + hostname + check.Path
		log.Printf("Checking %s...", url)

		status, body, err := client.Fetch(url)
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
		}

		if status != check.Status {
			msg := fmt.Sprintf("Invalid HTTP Response Status %d", status)
			messager.SendMessage(status, url, msg)
			continue
		}

		if check.Regex != "" {
			log.Println("Checking regex")
			re, err := regexp.Compile(check.Regex)
			if err != nil {
				log.Fatal(err)
			}

			matches := re.MatchString(body)
			if !matches {
				messager.SendMessage(status, url, "HTTP Response Body Error")
				continue
			}
		}

		if check.XMLSitemap {
			var sitemapUrls XMLSitemap
			err := xml.Unmarshal([]byte(body), &sitemapUrls)
			if err != nil {
				log.Fatal(err)
			}

			for _, xmlUrl := range sitemapUrls.URL {
				fmt.Printf("Checking %s... ", xmlUrl.Location)
				status, _, err := client.Fetch(xmlUrl.Location)
				if err != nil {
					log.Printf("Error: %s\n", err.Error())
				}

				if status != 200 {
					msg := fmt.Sprintf("Invalid HTTP Response Status %d", status)
					messager.SendMessage(status, url, msg)
					continue
				}
			}
		}

		log.Printf("%s Good\n", url)
	}
}
