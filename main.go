package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/bythepixel/urlchecker/pkg/config"
	"github.com/bythepixel/urlchecker/pkg/slack"
)

type HealthCheck struct {
	// Path to check
	Path string `json:"path"`

	// Regex used to check the body of the response
	Regex string `json:"regex"`

	// Status code expected from URL
	Status int `json:"status"`
}

func fetch(url string) (int, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}

	return resp.StatusCode, string(body), nil
}

func main() {
	// A Slack Webhook must be specified as an environment variable.
	webhook := os.Getenv(config.EnvSlackWebhook)
	if webhook == "" {
		log.Fatalf("Missing '%s' Environment Variable", config.EnvSlackWebhook)
	}
	slack := slack.SlackClient{
		Webhook: webhook,
	}

	// A filename must be specified for the program to read.
	var filename string
	flag.StringVar(&filename, "filename", "", "JSON File with paths")

	// A hostname must be specified since this may be used in different
	// environments.
	var hostname string
	flag.StringVar(&hostname, "hostname", "", "Hostname of website")

	var protocol string
	flag.StringVar(&protocol, "protocol", "https", "Protocol to use")

	flag.Parse()
	if filename == "" {
		log.Fatal("Missing filename flag")
	}
	if hostname == "" {
		log.Fatal("Missing hostname flag")
	}

	// Attempt to read the file specified.
	log.Printf("Reading %s...\n", filename)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Attempt to parse the file content as JSON.
	var urls []HealthCheck
	err = json.Unmarshal(bytes, &urls)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Loop through URLs and check each one.
	for _, check := range urls {
		url := protocol + "://" + hostname + check.Path
		log.Printf("Checking %s...\n", url)

		status, body, err := fetch(url)
		if err != nil {
			// Log the error and keep going.
			log.Printf("Error: %s\n", err.Error())
		}

		if status != check.Status {
			// Log the invalid response, send it to slack, then move onto the
			// next URL. We want to crawl every URL, so we don't exit if a URL
			// returns an incorrect response.
			msg := fmt.Sprintf("Invalid HTTP Response Status %d", status)
			slack.SendMessage(status, url, msg)
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
				log.Println("HTTP Response Body Error")
				slack.SendMessage(status, url, "HTTP Response Body Error")
				continue
			}
		}

		log.Println("Good")
	}
}
