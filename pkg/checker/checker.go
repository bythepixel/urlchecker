package checker

import (
	"encoding/json"
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

	// Loop through URLs and check each one.
	for _, check := range urls {
		url := protocol + "://" + hostname + check.Path
		log.Printf("Checking %s...\n", url)

		status, body, err := client.Fetch(url)
		if err != nil {
			// Log the error and keep going.
			log.Printf("Error: %s\n", err.Error())
		}

		if status != check.Status {
			// Log the invalid response, send it to slack, then move onto the
			// next URL. We want to crawl every URL, so we don't exit if a URL
			// returns an incorrect response.
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
				// log.Println("HTTP Response Body Error")
				messager.SendMessage(status, url, "HTTP Response Body Error")
				continue
			}
		}
	}
}
