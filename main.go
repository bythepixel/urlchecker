package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	EnvSlackWebhook = "SLACK_WEBHOOK"
	EnvGithubRepo   = "GITHUB_REPOSITORY"
)

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
func (c SlackClient) SendMessage(status int, message string) {
	repo := os.Getenv(EnvGithubRepo)
	msg := fmt.Sprintf("Repository: %s URL: %s responded with %d", repo, message, status)

	pl := SlackWebhookPayload{
		Text: msg,
	}

	jsonPayload, _ := json.Marshal(pl)

	_, err := http.Post(c.Webhook, "application/json", bytes.NewBuffer(jsonPayload))

	if err != nil {
		log.Fatal(err)
	}
}

func fetch(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
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
	var filename string
	flag.StringVar(&filename, "filename", "", "JSON File with URLs")
	flag.Parse()
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
	var urls []string
	err = json.Unmarshal(bytes, &urls)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Loop through URLs and check each one.
	for _, url := range urls {
		log.Printf("Checking %s...\n", url)
		status, err := fetch(url)
		if err != nil {
			// Log the error and keep going.
			log.Printf("Error: %s\n", err.Error())
		}

		if status != 200 {
			// Log the invalid response, send it to slack, then move onto the
			// next URL. We want to crawl every URL, so we don't exit if a URL
			// returns an incorrect response.
			log.Printf("Invalid HTTP Response Status %d\n", status)
			slack.SendMessage(status, url)
			continue
		}

		log.Println("Good")
	}
}
