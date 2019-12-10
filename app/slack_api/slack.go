package slackAPI

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"projects/slack-api/app/git"
	"projects/slack-api/config"
	"strings"

	"github.com/nlopes/slack"
)

type SlackApi struct {
	Client *slack.Client
	BotID  string
}

func (api *SlackApi) ListenOnEvent() {
	rtm := api.Client.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s>", api.BotID)) {
				break
			}

			attachment := slack.Attachment{
				Text:       "Hello! May I help you?",
				Color:      "#3AA3E3",
				CallbackID: "showdialog",
				Actions: []slack.AttachmentAction{
					{
						Name:  "indexIssue",
						Text:  "Show issues",
						Type:  "button",
						Value: "indexIssue",
					}, {
						Name:  "createIssue",
						Text:  "Create issue",
						Type:  "button",
						Value: "createIssue",
					},
				},
			}

			options := slack.MsgOptionAttachments(attachment)

			api.Client.PostMessage(ev.Channel, options)
		}
	}
}

func (api *SlackApi) interactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("invalid method: ", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("read request body failed: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Println("Failed to unescape request body: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var interaction slack.InteractionCallback

	if err := json.Unmarshal([]byte(jsonStr), &interaction); err != nil {
		log.Println("json unmarshal message failed: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if interaction.Type == slack.InteractionTypeInteractionMessage {
		switch interaction.ActionCallback.AttachmentActions[0].Value {
		case "createIssue":
			user := GetUserInfo(interaction.User.ID)
			if !user.Ok {
				log.Print("failed to get user info")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if err := api.Client.OpenDialogContext(context.TODO(), interaction.TriggerID, MakeDialog()); err != nil {
				log.Print("open dialog failed: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			text := fmt.Sprintf(":pencil2: Edit by @%s", user.User.Profile.DisplayName)
			overwriteMessage(w, interaction, text)
		}

	} else if interaction.Type == slack.InteractionTypeDialogSubmission {
		issue := git.MakeIssue(
			interaction.Submission["issueTitle"],
			interaction.Submission["issueContents"],
			interaction.Submission["screenShot"],
		)

		jsonValue, err := json.Marshal(issue)
		if err != nil {
			log.Println("marshal issue failed: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = git.PostIssue(bytes.NewBuffer(jsonValue), interaction.Submission["repository"])
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			attachment := slack.Attachment{
				Text:       ":x: Failed to post issue",
				Color:      "#3AA3E3",
				CallbackID: "showdialog",
			}
			options := slack.MsgOptionAttachments(attachment)
			api.Client.PostMessage(interaction.Channel.ID, options)
			return
		}

		attachment := slack.Attachment{
			Text:       ":white_check_mark: Completed!",
			Color:      "#3AA3E3",
			CallbackID: "showdialog",
		}
		options := slack.MsgOptionAttachments(attachment)
		api.Client.PostMessage(interaction.Channel.ID, options)

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "")
	}
}

func overwriteMessage(w http.ResponseWriter, interaction slack.InteractionCallback, text string) {
	originalMessage := interaction.OriginalMessage
	originalMessage.ReplaceOriginal = true
	originalMessage.Text = text
	originalMessage.Attachments = []slack.Attachment{}
	w.Header().Add("Content-type", "application/json")
	json.NewEncoder(w).Encode(&originalMessage)
}

func StartApiServer(api SlackApi) error {
	http.HandleFunc("/interaction", api.interactionHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), nil)
}
