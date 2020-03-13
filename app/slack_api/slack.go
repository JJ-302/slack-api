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
			if ev.Channel != config.Config.ChickTag {
				break
			}

			if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s>", api.BotID)) {
				break
			}

			msg := strings.Split(strings.TrimSpace(ev.Msg.Text), " ")[1:]
			if len(msg) != 0 && msg[0] == "issue" {
				attachment := slack.Attachment{
					Text:       "こんにちは！",
					Color:      "#2c2d30",
					CallbackID: "createIssue",
					Actions: []slack.AttachmentAction{
						{
							Name:  "createIssue",
							Text:  "Issueを作成する",
							Type:  "button",
							Style: "primary",
							Value: "createIssue",
						}, {
							Name:  "cancel",
							Text:  "キャンセル",
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
					Text:       "こんにちは！",
					Color:      "#2c2d30",
					CallbackID: "registerToken",
					Actions: []slack.AttachmentAction{
						{
							Name:  "registerToken",
							Text:  "トークンを登録する",
							Type:  "button",
							Style: "primary",
							Value: "registerToken",
						}, {
							Name:  "cancel",
							Text:  "キャンセル",
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
			text := fmt.Sprintf(":pencil2: @%s さんが編集中です。", user.User.Profile.DisplayName)
			overwriteMessage(w, interaction, text)

		case "registerToken":
			if err := api.Client.OpenDialogContext(context.TODO(), interaction.TriggerID, MakeTokenDialog()); err != nil {
				log.Println("open dialog failed: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			text := fmt.Sprintf(":pencil2: Edit by @%s", user.User.Profile.DisplayName)
			overwriteMessage(w, interaction, text)

		case "joinReach":
			if user.verify() {
				text := ":white_check_mark: 認証が完了しました！"
				overwriteMessage(w, interaction, text)
			} else {
				text := ":x: 認証に失敗しました。"
				overwriteMessage(w, interaction, text)
			}

		case "cancel":
			text := fmt.Sprintf(":x: @%s さんがキャンセルしました", user.User.Profile.DisplayName)
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
			var token git.Token
			token.Get(interaction.User.ID)

			var responseIssue git.ResponseIssue
			err = responseIssue.PostIssue(bytes.NewBuffer(jsonValue), interaction.Submission["repository"], token.Token)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				attachment := slack.Attachment{
					Text:       ":x: Issueの投稿に失敗しました。",
					Color:      "#2c2d30",
					CallbackID: "showdialog",
				}
				options := slack.MsgOptionAttachments(attachment)
				api.Client.PostMessage(interaction.Channel.ID, options)
				return
			}

			text := fmt.Sprintf(
				":white_check_mark: Issueを作成しました！\n\n%s\n%s", responseIssue.Title, responseIssue.HTMLURL)

			attachment := slack.Attachment{
				Text:       text,
				Color:      "#2c2d30",
				CallbackID: "showdialog",
			}
			options := slack.MsgOptionAttachments(attachment)
			api.Client.PostMessage(interaction.Channel.ID, options)

		case "registerToken":
			token := git.MakeToken(interaction.Submission["token"])
			err := token.Save(interaction.User.ID)
			if err != nil {
				log.Printf("Failed to save token: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				attachment := slack.Attachment{
					Text:       ":x: トークンの登録に失敗しました。",
					Color:      "#2c2d30",
					CallbackID: "registertoken",
				}
				options := slack.MsgOptionAttachments(attachment)
				api.Client.PostMessage(interaction.Channel.ID, options)
				return
			}

			attachment := slack.Attachment{
				Text:       ":white_check_mark: トークンを登録しました！",
				Color:      "#2c2d30",
				CallbackID: "showdialog",
			}
			options := slack.MsgOptionAttachments(attachment)
			api.Client.PostMessage(interaction.Channel.ID, options)
		}

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

func (api *SlackApi) messageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("invalid method: ", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeResponse(w, false)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("read request body failed: ", err)
		writeResponse(w, false)
		return
	}

	var user SlackUser
	if err = json.Unmarshal(body, &user); err != nil {
		log.Println("json unmarshal slack user failed: ", err)
		writeResponse(w, false)
		return
	}

	user.lookupUserByEmail()

	if user.Ok {
		attachment := slack.Attachment{
			Text:       "こんにちは！Reachへようこそ！",
			Color:      "#2c2d30",
			CallbackID: "joinReach",
			Actions: []slack.AttachmentAction{
				{
					Name:  "confirm",
					Text:  "参加する",
					Type:  "button",
					Style: "primary",
					Value: "joinReach",
				}, {
					Name:  "cancel",
					Text:  "キャンセル",
					Type:  "button",
					Style: "danger",
					Value: "cancel",
				},
			},
		}
		options := slack.MsgOptionAttachments(attachment)
		api.Client.PostMessage(user.User.ID, options)
		writeResponse(w, true)
	} else {
		log.Println("slack user does not exist.")
		writeResponse(w, false)
	}
}

func writeResponse(w http.ResponseWriter, result bool) {
	jsonValue, _ := json.Marshal(map[string]bool{"result": result})
	w.Write(jsonValue)
}

func StartApiServer(api SlackApi) error {
	http.HandleFunc("/interaction", api.interactionHandler)
	http.HandleFunc("/message", api.messageHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), nil)
}
