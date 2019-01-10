package webhook

import (
	"gitlack/resource/gitlab"
	"gitlack/resource/slack"
	"gitlack/store"
)

type Webhook interface {
	MergeRequestEvent([]byte)
	TagPushEvent([]byte)
	IssuesEvent([]byte)
	CommentsEvent([]byte)
}

type hook struct {
	db store.Store
	g  gitlab.GitLab
	s  slack.Slack
}

func NewWebhook(db store.Store, g gitlab.GitLab, s slack.Slack) Webhook {
	return &hook{
		db: db,
		g:  g,
		s:  s,
	}
}
