package slack

import (
	"fmt"
	"gitlack/resource"

	"github.com/urfave/cli"
)

type Slack interface {
	GetUser() ([]*SlackUser, error)
	PostSlackMessage(string, string, *Attachment, ...string) (*MessageResponse, error)
}

type slack struct {
	tool       resource.Util
	SlackAPI   string
	SlackToken string
}

func NewSlack(c *cli.Context) Slack {
	return &slack{
		tool:       resource.NewUtil(),
		SlackAPI:   fmt.Sprintf("%v://%v/api", c.String("slack-scheme"), c.String("slack-domain")),
		SlackToken: c.String("slack-token"),
	}
}
