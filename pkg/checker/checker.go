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

	for _, check := range urls {
		url := protocol + "://" + hostname + check.Path
		fmt.Printf("Checking %s... ", url)

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

		fmt.Println("Good")
	}
}
