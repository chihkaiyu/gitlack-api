package slack

import (
	"fmt"
	"gitlack/model"
	"gitlack/resource"

	"github.com/urfave/cli"
)

type Slack interface {
	GetUser() ([]*SlackUser, error)
	PostSlackMessage(string, string, *model.User, *Attachment, ...string) (*MessageResponse, error)
}

type slack struct {
	client     resource.Client
	SlackAPI   string
	SlackToken string
}

func NewSlack(c *cli.Context) Slack {
	return &slack{
		client:     resource.NewClient(),
		SlackAPI:   fmt.Sprintf("%v://%v/api", c.String("slack-scheme"), c.String("slack-domain")),
		SlackToken: c.String("slack-token"),
	}
}
