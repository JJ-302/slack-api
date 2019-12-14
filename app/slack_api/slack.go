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
	"projects/slack-api/app/db"
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
			if ev.Channel != config.Config.ChickTag {
				break
			}

			if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s>", api.BotID)) {
				break
			}

			msg := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
			if len(msg) != 0 && msg[0] == "issue" {
				attachment := slack.Attachment{
					Text:       "Hello! May I help you?",
					Color:      "#2c2d30",
					CallbackID: "createIssue",
					Actions: []slack.AttachmentAction{
						{
							Name:  "createIssue",
							Text:  "Create issue",
							Type:  "button",
							Style: "primary",
							Value: "createIssue",
						}, {
							Name:  "cancel",
							Text:  "Cancel",
							Type:  "button",
							Style: "danger",
							Value: "cancel",
						},
					},
				}
				options := slack.MsgOptionAttachments(attachment)
				api.Client.PostMessage(ev.Channel, options)

			} else if len(msg) != 0 && msg[0] == "token" {
				attachment := slack.Attachment{
					Text:       "Hello! May I help you?",
					Color:      "#2c2d30",
					CallbackID: "registerToken",
					Actions: []slack.AttachmentAction{
						{
							Name:  "registerToken",
							Text:  "Register token",
							Type:  "button",
							Style: "primary",
							Value: "registerToken",
						}, {
							Name:  "cancel",
							Text:  "Cancel",
							Type:  "button",
							Style: "danger",
							Value: "cancel",
						},
					},
				}
				options := slack.MsgOptionAttachments(attachment)
				api.Client.PostMessage(ev.Channel, options)
			}
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
		user := GetUserInfo(interaction.User.ID)
		if !user.Ok {
			log.Println("failed to get user info")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch interaction.ActionCallback.AttachmentActions[0].Value {
		case "createIssue":
			uid := interaction.User.ID
			if err := api.Client.OpenDialogContext(context.TODO(), interaction.TriggerID, MakeIssueDialog(uid)); err != nil {
				log.Println("open dialog failed: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			text := fmt.Sprintf(":pencil2: Edit by @%s", user.User.Profile.DisplayName)
			overwriteMessage(w, interaction, text)

		case "registerToken":
			if err := api.Client.OpenDialogContext(context.TODO(), interaction.TriggerID, MakeTokenDialog()); err != nil {
				log.Println("open dialog failed: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			text := fmt.Sprintf(":pencil2: Edit by @%s", user.User.Profile.DisplayName)
			overwriteMessage(w, interaction, text)

		case "cancel":
			text := fmt.Sprintf(":x: Canceled by @%s", user.User.Profile.DisplayName)
			overwriteMessage(w, interaction, text)
		}

	} else if interaction.Type == slack.InteractionTypeDialogSubmission {
		switch interaction.CallbackID {
		case "createIssue":
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
			var token db.Token
			token.Get(interaction.User.ID)

			err = git.PostIssue(bytes.NewBuffer(jsonValue), interaction.Submission["repository"], token.Token)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				attachment := slack.Attachment{
					Text:       ":x: Failed to post issue",
					Color:      "#2c2d30",
					CallbackID: "showdialog",
				}
				options := slack.MsgOptionAttachments(attachment)
				api.Client.PostMessage(interaction.Channel.ID, options)
				return
			}

		case "registerToken":
			token := db.New(interaction.Submission["token"])
			err := token.Save(interaction.User.ID)
			if err != nil {
				log.Printf("Failed to save token: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				attachment := slack.Attachment{
					Text:       ":x: Failed to post issue",
					Color:      "#2c2d30",
					CallbackID: "registertoken",
				}
				options := slack.MsgOptionAttachments(attachment)
				api.Client.PostMessage(interaction.Channel.ID, options)
				return
			}
		}

		attachment := slack.Attachment{
			Text:       ":white_check_mark: Completed!",
			Color:      "#2c2d30",
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
