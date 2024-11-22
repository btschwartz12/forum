package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/btschwartz12/forum/repo"
	"github.com/btschwartz12/slackalerts"
	"go.uber.org/zap"
)

func SendUserCreatedAlert(slackWebhookUrl string, logger *zap.SugaredLogger, user *repo.User) {
	blocks := getBlocksForUser(user)
	title := "New User Created"
	err := slackalerts.SendAlert(context.Background(), slackWebhookUrl, title, blocks)
	if err != nil {
		logger.Errorw("error sending slack alert", "error", err)
	}
}

func SendPostCreatedAlert(slackWebhookUrl string, logger *zap.SugaredLogger, post *repo.Post) {
	blocks := getBlocksForPost(post)
	title := "New Post Created"
	err := slackalerts.SendAlert(context.Background(), slackWebhookUrl, title, blocks)
	if err != nil {
		logger.Errorw("error sending slack alert", "error", err)
	}
}

func getBlocksForUser(user *repo.User) []slackalerts.Block {
	now := repo.EstTime{Time: time.Now()}
	blocks := []slackalerts.Block{
		{
			Type: "header",
			Text: &slackalerts.Element{
				Type:  "plain_text",
				Text:  "New User Created 🎉",
				Emoji: true,
			},
		},
		{
			Type: "context",
			Elements: []slackalerts.Element{
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("created at %s", now),
				},
			},
		},
		{
			Type: "context",
			Elements: []slackalerts.Element{
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("username: `%s`", user.Username),
				},
			},
		},
	}
	return blocks
}

func getBlocksForPost(post *repo.Post) []slackalerts.Block {
	blocks := []slackalerts.Block{
		{
			Type: "header",
			Text: &slackalerts.Element{
				Type:  "plain_text",
				Text:  "New Post Created 🎉",
				Emoji: true,
			},
		},
		{
			Type: "context",
			Elements: []slackalerts.Element{
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("`%s` posted at %s from ip=%s", post.Author.Username, post.Timestamp, post.Ip),
				},
			},
		},
		{
			Type: "context",
			Elements: []slackalerts.Element{
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("content: %s", post.Content),
				},
			},
		},
	}
	return blocks
}
