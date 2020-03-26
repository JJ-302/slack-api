package slackAPI

import (
	"projects/slack-api/app/git"

	"github.com/nlopes/slack"
)

func MakeIssueDialog(uid string) slack.Dialog {
	repos := git.GetRepos(uid)
	var dialogSelection []slack.DialogSelectOption
	for _, repo := range *repos {
		selection := slack.DialogSelectOption{
			Label: repo.Name,
			Value: repo.FullName,
		}
		dialogSelection = append(dialogSelection, selection)
	}

	dialog := slack.Dialog{
		Title:       "Create issue",
		SubmitLabel: "Submit",
		CallbackID:  "createIssue",
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
				Placeholder: "内容を入力してください（マークダウン可）",
			},
			slack.DialogInput{
				Label:       "ScreenShot",
				Type:        slack.InputTypeTextArea,
				Name:        "screenShot",
				Optional:    true,
				Placeholder: "画像のURLをマークダウンで貼り付けてください",
			},
		},
	}
	return dialog
}

func MakeTokenDialog() slack.Dialog {
	dialog := slack.Dialog{
		Title:       "Register token",
		SubmitLabel: "Submit",
		CallbackID:  "registerToken",
		Elements: []slack.DialogElement{
			slack.DialogInput{
				Label:       "Token",
				Type:        slack.InputTypeText,
				Name:        "token",
				Placeholder: "トークンを入力してください",
			},
		},
	}
	return dialog
}

func MakeReleaseDialog() slack.Dialog {
	dialog := slack.Dialog{
		Title:       "Release",
		SubmitLabel: "Submit",
		CallbackID:  "postRelease",
		Elements: []slack.DialogElement{
			slack.DialogInputSelect{
				DialogInput: slack.DialogInput{
					Label:       "Release channel",
					Type:        slack.InputTypeSelect,
					Name:        "releaseChannel",
					Placeholder: "Release channelを選択してください",
				},
				Options: []slack.DialogSelectOption{
					slack.DialogSelectOption{
						Label: "Production",
						Value: "Production",
					},
					slack.DialogSelectOption{
						Label: "Staging",
						Value: "Staging",
					},
				},
			},
			slack.DialogInputSelect{
				DialogInput: slack.DialogInput{
					Label:       "Platform",
					Type:        slack.InputTypeSelect,
					Name:        "platform",
					Placeholder: "Platformを選択してください",
				},
				Options: []slack.DialogSelectOption{
					slack.DialogSelectOption{
						Label: "iOS",
						Value: "iOS",
					},
					slack.DialogSelectOption{
						Label: "Android",
						Value: "Android",
					},
				},
			},
			slack.DialogInput{
				Label:       "Version",
				Type:        slack.InputTypeText,
				Name:        "version",
				Placeholder: "Versionを入力してください",
			},
			slack.DialogInput{
				Label:       "Release note",
				Type:        slack.InputTypeTextArea,
				Name:        "releaseNote",
				Placeholder: "Release noteを入力してください",
			},
		},
	}
	return dialog
}
