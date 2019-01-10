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
}

type gitlab struct {
	tool         resource.Util
	GitLabDomain string
	GitLabAPI    string
	GitLabToken  string
}

func NewGitLab(c *cli.Context) GitLab {
	return &gitlab{
		tool:         resource.NewUtil(),
		GitLabDomain: c.String("gitlab-domain"),
		GitLabAPI:    fmt.Sprintf("%v://%v/api/v4", c.String("gitlab-scheme"), c.String("gitlab-domain")),
		GitLabToken:  c.String("gitlab-token"),
	}
}
