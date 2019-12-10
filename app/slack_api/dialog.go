package slackAPI

import (
	"projects/slack-api/app/git"

	"github.com/nlopes/slack"
)

func MakeDialog() slack.Dialog {
	repos := git.GetRepos()
	var dialogSelection []slack.DialogSelectOption
	for _, repo := range *repos {
		selection := slack.DialogSelectOption{
			Label: repo.Name,
			Value: repo.URL,
		}
		dialogSelection = append(dialogSelection, selection)
	}

	dialog := slack.Dialog{
		Title:       "Create issue",
		SubmitLabel: "Submit",
		CallbackID:  "coffee_order_form",
		Elements: []slack.DialogElement{
			slack.DialogInputSelect{
				DialogInput: slack.DialogInput{
					Label:       "Repository",
					Type:        slack.InputTypeSelect,
					Name:        "repository",
					Placeholder: "リポジトリを選択してください",
				},
				Options: dialogSelection,
			},
			slack.DialogInput{
				Label:       "Title",
				Type:        slack.InputTypeText,
				Name:        "issueTitle",
				Placeholder: "タイトルを入力してください",
			},
			slack.DialogInput{
				Label:       "Contents",
				Type:        slack.InputTypeTextArea,
				Name:        "issueContents",
				Placeholder: "内容を入力してください",
			},
			slack.DialogInput{
				Label:       "ScreenShot",
				Type:        slack.InputTypeTextArea,
				Name:        "screenShot",
				Optional:    true,
				Placeholder: "マークダウン形式で貼り付けてください",
			},
		},
	}
	return dialog
}
