package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlack/model"

	"github.com/sirupsen/logrus"
)

// GitLabProject is the response of getting GitLab project list
type GitLabProject struct {
	Name string `json:"path_with_namespace"`
	ID   int    `json:"id"`
}

func (g *gitlab) GetProject() ([]*model.Project, error) {
	url := g.GitLabAPI + "/projects"
	params := map[string]string{
		"private_token": g.GitLabToken,
		"per_page":      "100",
		"archived":      "false",
		"simple":        "true",
	}

	var allProjects []*model.Project
	// run at most 100 times for preventing from infinite loop
	for i := 0; i < 100; i++ {
		res, err := g.tool.Request(http.MethodGet, url, nil, params, nil)
		if err != nil {
			return nil, err
		}
		body, e := ioutil.ReadAll(res.Body)
		if e != nil {
			logrus.Errorln(e)
			return nil, e
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			errMsg := fmt.Sprintf("GitLab error: %v", string(body))
			logrus.Errorln(errMsg)
			return nil, errors.New(errMsg)
		}

		var project []GitLabProject
		err = json.Unmarshal(body, &project)
		if err != nil {
			logrus.Errorln(err)
			return nil, err
		}
		for _, p := range project {
			allProjects = append(allProjects, &model.Project{
				ID:   p.ID,
				Name: p.Name,
			})
		}

		// check next page
		if res.Header["X-Next-Page"][0] == "" {
			break
		}
		params["page"] = res.Header["X-Next-Page"][0]
	}
	return allProjects, nil
}
