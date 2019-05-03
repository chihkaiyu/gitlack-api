package gitlab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

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
	CommitterEmail string   `json:"committer_email"`
	LastPipeline   Pipeline `json:"last_pipeline"`
}

type Pipeline struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

type Compare struct {
	Commits []*Commit `json:"commits"`
}

func (g *gitlab) GetTagList(id int) ([]*Tag, error) {
	url := g.GitLabAPI + fmt.Sprintf("/projects/%v/repository/tags", id)
	params := map[string]string{
		"private_token": g.GitLabToken,
	}
	res, err := g.client.Get(url, nil, params, nil)
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
		err := fmt.Errorf("Invalid GitLab API error: %v", string(body))
		logrus.Errorln(err)
		return nil, err
	}
	var tagList []*Tag
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	return tagList, nil
}
