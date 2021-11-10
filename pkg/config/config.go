package config

import "os"

const (
	EnvGithubRepo   = "GITHUB_REPOSITORY"
	EnvSlackWebhook = "SLACK_WEBHOOK"
)

var Debug = false

func init() {
	if os.Getenv("DEBUG") != "" {
		Debug = true
	}
}
