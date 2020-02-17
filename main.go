package main

import (
	"log"
	slackAPI "projects/slack-api/app/slack_api"
	"projects/slack-api/config"

	"github.com/nlopes/slack"
)

func main() {
	api := slackAPI.SlackApi{
		Client: slack.New(config.Config.Token),
		BotID:  config.Config.BotID,
	}

	go api.ListenOnEvent()

	log.Println(slackAPI.StartApiServer(api))
}
