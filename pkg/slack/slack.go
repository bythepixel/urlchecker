package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bythepixel/urlchecker/pkg/config"
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
func (c SlackClient) SendMessage(status int, url string, message string) {
	repo := os.Getenv(config.EnvGithubRepo)
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
