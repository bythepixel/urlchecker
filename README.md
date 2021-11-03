# urlchecker

Below is an example YAML file for this action.

```yaml
name: Check URLs

on:
  push:
    branches:
      - '*'

env:
  SLACK_WEBHOOK: ${{ secrets.SLACK_WEBHOOK }}

jobs:
  check-urls:
    runs-on: ubuntu-latest
    name: Checks URLs from JSON file
    steps:
      - name: Checkout
        uses: actions/checkout@v2
      - name: Check URLs
        uses: bythepixel/urlchecker
        args:
          - ./urls.json
```

## Description

This GitHub Action reads a JSON file, crawls the URLs, and checks the resposnes.

## Requirements

* A `SLACK_WEBHOOK` URL to send a message when something goes wrong
* A JSON file of URLs in your repository that uses the following structure

## JSON File

```json
[
    {
        "url": "https://postman-echo.com/status/200",
        "status": 200
    },
    {
        "url": "https://postman-echo.com/status/200",
        "status": 200
    },
    {
        "url": "https://postman-echo.com/status/200",
        "status": 200,
        "regex": "200"
    }
]
```

View the files in the [json](json) folder to see more examples. See the Golang
[regexp][1] package for additional information on supported regular expressions.

## Environment Variables

This Action uses these environment variables

* `SLACK_WEBHOOK` is one you need to provide
* `GITHUB_REPOSITORY` is provided by GitHub

[1]: https://pkg.go.dev/regexp
