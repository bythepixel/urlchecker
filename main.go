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

const EnvSlackWebhook = "SLACK_WEBHOOK"

type payload struct {
	Text string `json:"text"`
}

type SlackClient struct {
	Webhook string
}

func (c SlackClient) SendMessage(status int, message string) {
	msg := fmt.Sprintf("%s responded with %d", message, status)

	pl := payload{
		Text: msg,
	}

	jsonPayload, _ := json.Marshal(pl)

	_, err := http.Post(c.Webhook, "application/json",
		bytes.NewBuffer(jsonPayload))

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
	log.Println("Starting...")

	webhook := os.Getenv(EnvSlackWebhook)
	if webhook == "" {
		log.Fatalf("Missing '%s' Environment Variable", EnvSlackWebhook)
	}

	var filename string
	flag.StringVar(&filename, "filename", "", "JSON File with URLs")
	flag.Parse()
	if filename == "" {
		log.Fatal("Missing filename flag")
	}

	slack := SlackClient{
		Webhook: webhook,
	}

	log.Printf("Reading %s...\n", filename)

	bytes, _ := ioutil.ReadFile(filename)

	var urls []string
	json.Unmarshal(bytes, &urls)

	for _, url := range urls {
		log.Printf("Checking %s...\n", url)
		status, err := fetch(url)
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
		}

		if status != 200 {
			log.Printf("Invalid HTTP Response Status %d\n", status)
			slack.SendMessage(status, url)
			continue
		}

		log.Println("Good")
	}
}
