package git

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"projects/slack-api/config"
	"time"
)

var issueFormat = "### 内容\n\n%s\n\n### スクリーンショット\n\n%s"

type Issue struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type ResponseIssue struct {
	HTMLURL    string    `json:"html_url"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"created_at"`
	Repository Repository
}

func MakeIssue(title, body, screenshot string) *Issue {
	content := fmt.Sprintf(issueFormat, body, screenshot)
	issue := Issue{
		Title: title,
		Body:  content,
	}
	return &issue
}

func (issue *ResponseIssue) PostIssue(body io.Reader, url, token string) error {
	req, err := http.NewRequest(http.MethodPost, config.Config.RequestURL+"/repos/"+url+"/issues", body)
	if err != nil {
		log.Println("failed to build request for post issue: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+token)

	client := http.DefaultClient

	response, err := client.Do(req)
	if err != nil {
		log.Println("failed to request: ", err)
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("failed to create issue")
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("read response body failed: ", err)
		return err
	}

	if err = json.Unmarshal(responseBody, issue); err != nil {
		log.Println("json unmarshal message failed: ", err)
		return err
	}

	return nil
}
