package handler

import (
	"fmt"
	"gitlack/model"
	"gitlack/resource/gitlab"
	"gitlack/resource/slack"
	"strings"

	"github.com/stretchr/testify/mock"

	mGitLab "gitlack/resource/gitlab/mocks"
	mSlack "gitlack/resource/slack/mocks"
	mDB "gitlack/store/mocks"
)

func getGitLabUser(start, count int) []*gitlab.GitLabUser {
	var gitlabUsers []*gitlab.GitLabUser
	for i := start; i < start+count; i++ {
		g := &gitlab.GitLabUser{
			ID:    i,
			Email: fmt.Sprintf("fake-%v@fake.com", i),
			Name:  fmt.Sprintf("fake-%v", i),
		}
		gitlabUsers = append(gitlabUsers, g)
	}
	return gitlabUsers
}

func getSlackuser(start, count int) []*slack.SlackUser {
	var slackUsers []*slack.SlackUser
	for i := start; i < start+count; i++ {
		s := &slack.SlackUser{
			ID:        fmt.Sprintf("fake-%v", i),
			Email:     fmt.Sprintf("fake-%v@fake.com", i),
			AvatarURL: fmt.Sprintf("http://fake.com/fake-%v.jpg", i),
		}
		slackUsers = append(slackUsers, s)
	}

	return slackUsers
}

func getUser(gitlabUser []*gitlab.GitLabUser, slackUser []*slack.SlackUser) []*model.User {
	combinedUser := make(map[string]*model.User)
	for _, g := range gitlabUser {
		email := strings.Split(g.Email, "@")[0]
		u := &model.User{
			Email:    email,
			GitLabID: g.ID,
			Name:     g.Name,
		}
		combinedUser[g.Email] = u
	}

	for _, s := range slackUser {
		if u, exist := combinedUser[s.Email]; exist {
			u.SlackID = s.ID
			u.AvatarURL = s.AvatarURL
			combinedUser[s.Email] = u
		}
	}

	var res []*model.User
	for _, u := range combinedUser {
		res = append(res, u)
	}

	return res
}

func getStubGetUserGitLab(gitlabUser []*gitlab.GitLabUser) *mGitLab.GitLab {
	g := &mGitLab.GitLab{}
	g.On("GetUser").Return(gitlabUser, nil)

	return g
}

func getStubGetUserSlack(slackUser []*slack.SlackUser) *mSlack.Slack {
	s := &mSlack.Slack{}
	s.On("GetUser").Return(slackUser, nil)

	return s
}

func getStubDB(methodName string, err error) *mDB.Store {
	db := &mDB.Store{}
	db.On(methodName, mock.Anything).Return(err)

	return db
}

func getRouter(db *mDB.Store, slack *mSlack.Slack, gitlab *mGitLab.GitLab) *router {
	return &router{
		db: db,
		s:  slack,
		g:  gitlab,
	}
}

func getProjects(count int) []*model.Project {
	var projects []*model.Project
	for i := 0; i < count; i++ {
		p := &model.Project{
			ID:   i,
			Name: fmt.Sprintf("fake-project-%v", i),
		}
		projects = append(projects, p)
	}

	return projects
}

func getStubGetProjectGitLab(project []*model.Project) *mGitLab.GitLab {
	g := &mGitLab.GitLab{}
	g.On("GetProject").Return(project, nil)

	return g
}
