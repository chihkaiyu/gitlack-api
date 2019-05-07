package gitlab

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gitlack/resource/mocks"
	"io/ioutil"
	"net/http"

	"github.com/stretchr/testify/mock"
)

func getGitLab(client *mocks.Client) *gitlab {
	return &gitlab{
		client: client,
	}
}

func getGitLabUserResponse(count int) ([]byte, []*GitLabUser) {
	var members []*GitLabUser
	for i := 0; i < count; i++ {
		m := &GitLabUser{
			ID:    i,
			Email: fmt.Sprintf("fake-%v@fake.com", i),
			Name:  fmt.Sprintf("fake-%v", i),
		}
		members = append(members, m)
	}

	res, _ := json.Marshal(members)
	return res, members
}

func getClient() *mocks.Client {
	return &mocks.Client{}
}

func getNextPageHeader(page string) map[string]string {
	return map[string]string{
		"X-Next-Page": page,
	}
}

func getResponse(stubByte []byte, statusCode int, header map[string]string) *http.Response {
	h := make(map[string][]string)
	for k, v := range header {
		h[k] = []string{v}
	}
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(stubByte)),
		Header:     h,
	}
}

func getGetClientWithResponse(stubByte []byte, statusCode int, page string) *mocks.Client {
	stubResponse := getResponse(stubByte, statusCode, getNextPageHeader(page))
	stubClient := getClient()
	stubClient.On(
		"Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(stubResponse, nil)

	return stubClient
}

func getGetClientWithError(errMsg string) *mocks.Client {
	stubClient := getClient()
	stubClient.On(
		"Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil, errors.New(errMsg))

	return stubClient
}

func getTagListResponse(count int) ([]byte, []*Tag) {
	var tagList []*Tag
	for i := 0; i < count; i++ {
		t := &Tag{
			Name: fmt.Sprintf("fake-tag-%v", i),
			ReleaseInfo: Release{
				Description: fmt.Sprintf("fake-description-%v", i),
			},
			CommitInfo: Commit{
				CommitterEmail: fmt.Sprintf("fake-%v@fake.com", i),
				LastPipeline: Pipeline{
					ID:     i,
					Status: fmt.Sprintf("fake-status-%v", i),
					WebURL: fmt.Sprintf("fake-web-url-%v", i),
				},
			},
		}

		tagList = append(tagList, t)
	}

	res, _ := json.Marshal(tagList)
	return res, tagList
}

func getCommitResponse(count int) ([]byte, Compare) {
	var compare Compare
	var commits []*Commit
	for i := 0; i < count; i++ {
		c := &Commit{
			CommitterEmail: fmt.Sprintf("fake-%v@fake.com", i),
			LastPipeline: Pipeline{
				ID:     i,
				Status: fmt.Sprintf("fake-status-%v", i),
				WebURL: fmt.Sprintf("fake-web-url-%v", i),
			},
		}

		commits = append(commits, c)
	}
	compare.Commits = commits

	res, _ := json.Marshal(compare)
	return res, compare
}

func getProjectResponse(count int) ([]byte, []*GitLabProject) {
	var gitlabProjects []*GitLabProject
	for i := 0; i < count; i++ {
		g := &GitLabProject{
			Name: fmt.Sprintf("fake-gitlab-project-%v", i),
			ID:   i,
		}
		gitlabProjects = append(gitlabProjects, g)
	}

	res, _ := json.Marshal(gitlabProjects)
	return res, gitlabProjects
}

func getCommit() ([]byte, *Commit) {
	c := &Commit{
		CommitterEmail: "fake-0@fake.com",
		LastPipeline: Pipeline{
			ID:     0,
			Status: "fake-status-0",
			WebURL: "fake-web-url-0",
		},
	}

	res, _ := json.Marshal(c)
	return res, c
}

func getListCommit(count int) ([]byte, []*Commit) {
	var commits []*Commit
	for i := 0; i < count; i++ {
		c := &Commit{
			CommitterEmail: fmt.Sprintf("fake-%v@fake.com", i),
			LastPipeline: Pipeline{
				ID:     i,
				Status: fmt.Sprintf("fake-status-%v", i),
				WebURL: fmt.Sprintf("fake-web-url-%v", i),
			},
		}
		commits = append(commits, c)
	}

	res, _ := json.Marshal(commits)
	return res, commits
}
