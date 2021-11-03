package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

const (
	EnvSlackWebhook = "SLACK_WEBHOOK"
	EnvGithubRepo   = "GITHUB_REPOSITORY"
	EnvFilename     = "INPUT_FILENAME"
)

type HealthCheck struct {
	// URL to check
	URL string `json:"url"`

	// Status code expected from URL
	Status int `json:"status"`

	// Regex used to check the body of the response
	Regex string `json:"regex"`
}

// SlackWebhookPayload represents the minimum required fields to send a webhook.
//
// https://api.slack.com/messaging/webhooks
type SlackWebhookPayload struct {
	Text string `json:"text"`
}

// SlackClient contains the Webhook URL.
type SlackClient struct {
	Webhook string
}

// SendMessage creates a SlackWebhookPayload and sends it to the Webhook URL.
func (c SlackClient) SendMessage(status int, url string, message string) {
	repo := os.Getenv(EnvGithubRepo)
	msg := fmt.Sprintf("Repository: %s URL: %s Message: %s", repo, url, message)

	pl := SlackWebhookPayload{
		Text: msg,
	}

	jsonPayload, _ := json.Marshal(pl)

	_, err := http.Post(c.Webhook, "application/json", bytes.NewBuffer(jsonPayload))

	if err != nil {
		log.Fatal(err)
	}
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
	webhook := os.Getenv(EnvSlackWebhook)
	if webhook == "" {
		log.Fatalf("Missing '%s' Environment Variable", EnvSlackWebhook)
	}
	slack := SlackClient{
		Webhook: webhook,
	}

	// A filename must be specified for the program to read.
	filename := os.Getenv(EnvFilename)
	// flag.StringVar(&filename, "filename", "", "JSON File with URLs")
	// flag.Parse()
	if filename == "" {
		log.Fatal("Missing filename flag")
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
		log.Printf("Checking %s...\n", check.URL)

		status, body, err := fetch(check.URL)
		if err != nil {
			// Log the error and keep going.
			log.Printf("Error: %s\n", err.Error())
		}

		if status != check.Status {
			// Log the invalid response, send it to slack, then move onto the
			// next URL. We want to crawl every URL, so we don't exit if a URL
			// returns an incorrect response.
			msg := fmt.Sprintf("Invalid HTTP Response Status %d", status)
			slack.SendMessage(status, check.URL, msg)
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
				slack.SendMessage(status, check.URL, "HTTP Response Body Error")
				continue
			}
		}

		log.Println("Good")
	}
}
