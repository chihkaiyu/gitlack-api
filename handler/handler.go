package handler

import (
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"

	"gitlack/handler/webhook"
	"gitlack/resource/gitlab"
	"gitlack/resource/slack"
	"gitlack/store"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
)

// Handler defines the interface that Gitlack needs to expose to outside
type Handler interface {
	GetProject(*gin.Context)
	UpdateProject(*gin.Context)
	WrapSyncProject(*gin.Context)
	SyncProject() error

	UpdateGroup(*gin.Context)

	GetUser(*gin.Context)
	UpdateUser(*gin.Context)
	WrapSyncUser(*gin.Context)
	SyncUser() error

	Webhook(*gin.Context)
}

type router struct {
	db   store.Store
	g    gitlab.GitLab
	s    slack.Slack
	hook webhook.Webhook
}

// NewHandler create a Handler
func NewHandler(c *cli.Context) Handler {
	db := store.NewStore(c)
	g := gitlab.NewGitLab(c)
	s := slack.NewSlack(c)
	h := webhook.NewWebhook(db, g, s)
	return &router{
		db:   db,
		g:    g,
		s:    s,
		hook: h,
	}
}

func (r *router) Webhook(c *gin.Context) {
	// GitLab doesn't care what you return to it
	// ref: https://docs.gitlab.com/ce/user/project/integrations/webhooks.html#webhook-endpoint-tips
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Errorln(err)
		c.JSON(http.StatusOK, gin.H{
			"ok": true,
		})
		return
	}
	switch event := c.GetHeader("X-Gitlab-Event"); event {
	case "Tag Push Hook":
		r.hook.TagPushEvent(body)
	case "Merge Request Hook":
		r.hook.MergeRequestEvent(body)
	case "Issue Hook":
		r.hook.IssuesEvent(body)
	case "Note Hook":
		r.hook.CommentsEvent(body)
	default:
		logrus.Infof("Event not supported: %v", event)
	}
	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}
