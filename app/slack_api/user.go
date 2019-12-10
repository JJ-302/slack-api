package slackAPI

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"projects/slack-api/config"
)

type SlackUser struct {
	Ok   bool `json:"ok"`
	User struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Profile struct {
			DisplayName string `json:"display_name"`
		} `json:"profile"`
	} `json:"user"`
}

func GetUserInfo(userID string) *SlackUser {
	requestURL := config.Config.SlackURL + "/users.info?token=" + config.Config.Token + "&user=" + userID
	response, err := http.Get(requestURL)
	if err != nil {
		log.Println("failed to request for get user info: ", err)
		return &SlackUser{}
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("read request body failed: ", err)
		return &SlackUser{}
	}

	var user SlackUser
	if err = json.Unmarshal(body, &user); err != nil {
		log.Println("json unmarshal slack user failed: ", err)
		return &SlackUser{}
	}

	return &user
}
