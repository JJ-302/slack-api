package git

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"projects/slack-api/config"
)

type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

func GetRepos(uid string) *[]Repository {
	url := config.Config.RequestURL
	req, err := http.NewRequest(http.MethodGet, url+"/orgs/Chick-Tag/repos", nil)
	if err != nil {
		log.Println("failed to build request for get repos: ", err)
		return nil
	}

	var token Token
	token.Get(uid)

	req.Header.Set("Authorization", "token "+token.Token)
	client := http.DefaultClient

	response, err := client.Do(req)
	if err != nil {
		log.Println("failed to request: ", err)
		return nil
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("read request body failed: ", err)
		return nil
	}

	var repos []Repository
	if err = json.Unmarshal(body, &repos); err != nil {
		log.Println("json unmarshal message failed: ", err)
		return nil
	}

	return &repos
}
