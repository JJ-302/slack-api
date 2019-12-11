package config

import (
	"log"
	"os"

	ini "gopkg.in/ini.v1"
)

type ConfigList struct {
	Port       int
	ChickTag   string
	Token      string
	SlackURL   string
	BotID      string
	RequestURL string
	ProjectID  string
}

var Config ConfigList

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		os.Exit(1)
	}

	Config = ConfigList{
		Port:       cfg.Section("web").Key("port").MustInt(),
		ChickTag:   cfg.Section("api").Key("chicktag").String(),
		Token:      cfg.Section("api").Key("token").String(),
		SlackURL:   cfg.Section("api").Key("url").String(),
		BotID:      cfg.Section("api").Key("botID").String(),
		RequestURL: cfg.Section("git").Key("url").String(),
		ProjectID:  cfg.Section("db").Key("projectID").String(),
	}
}
