package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Project struct {
	PathWithNamespace string `json:"path_with_namespace"`
	ID                int    `json:"id"`
	WebURL            string `json:"web_url"`
}

type Tag struct {
	Name        string  `json:"name"`
	ReleaseInfo Release `json:"release"`
	CommitInfo  Commit  `json:"commit"`
}

type Release struct {
	Description string `json:"description"`
}

type Commit struct {
	CommitterEmail string `json:"committer_email"`
}

type Compare struct {
	Commits []*Commit `json:"commits"`
}

func (g *gitlab) GetTagList(id int) ([]*Tag, error) {
	url := g.GitLabAPI + fmt.Sprintf("/projects/%v/repository/tags", id)
	params := map[string]string{
		"private_token": g.GitLabToken,
	}
	res, err := g.tool.Request(http.MethodGet, url, nil, params, nil)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		errMsg := fmt.Sprintf("GitLab error: %v", string(body))
		logrus.Errorln(errMsg)
		return nil, errors.New(errMsg)
	}
	var tagList []*Tag
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	return tagList, nil
}
