package gitlab

import (
	"fmt"

	"github.com/urfave/cli"
	"gitlack/model"
	"gitlack/resource"
)

type GitLab interface {
	GetProject() ([]*model.Project, error)
	GetUser() ([]*GitLabUser, error)
	GetTagList(int) ([]*Tag, error)
	GetSingleCommit(int, string) (*Commit, error)
}

type gitlab struct {
	client       resource.Client
	GitLabDomain string
	GitLabAPI    string
	GitLabToken  string
}

func NewGitLab(c *cli.Context) GitLab {
	return &gitlab{
		client:       resource.NewClient(),
		GitLabDomain: c.String("gitlab-domain"),
		GitLabAPI:    fmt.Sprintf("%v://%v/api/v4", c.String("gitlab-scheme"), c.String("gitlab-domain")),
		GitLabToken:  c.String("gitlab-token"),
	}
}
