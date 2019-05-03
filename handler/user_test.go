package handler

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlack/model"

	"gitlack/resource/gitlab"
	"gitlack/resource/slack"

	mGitLab "gitlack/resource/gitlab/mocks"
	mSlack "gitlack/resource/slack/mocks"
	mDB "gitlack/store/mocks"
)

func TestSyncUser(t *testing.T) {
	mockedDB := &mDB.Store{}
	mockedSlack := &mSlack.Slack{}
	mockedGitLab := &mGitLab.GitLab{}

	var mockedGitLabUser []*gitlab.GitLabUser
	var mockedSlackUser []*slack.SlackUser
	var mockedUser []*model.User
	for i := 0; i < 10; i++ {
		fakeEmail := fmt.Sprintf("fake-email-%d@fake.com", i)
		fakeName := fmt.Sprintf("fake-name-%d", i)
		fakeSlackID := fmt.Sprintf("fake-slack-id-%d", i)
		fakeAvatarURL := fmt.Sprintf("https://fake.com/fake-slack-avatar-%d.png", i)
		g := &gitlab.GitLabUser{ID: i, Email: fakeEmail, Name: fakeName}
		mockedGitLabUser = append(mockedGitLabUser, g)
		s := &slack.SlackUser{ID: fakeSlackID, Email: fakeEmail, AvatarURL: fakeAvatarURL}
		mockedSlackUser = append(mockedSlackUser, s)
		u := &model.User{
			Email:     strings.Split(fakeEmail, "@")[0],
			SlackID:   s.ID,
			GitLabID:  g.ID,
			Name:      g.Name,
			AvatarURL: s.AvatarURL,
		}
		mockedUser = append(mockedUser, u)
	}

	mockedGitLab.On("GetUser").Return(mockedGitLabUser, nil)
	mockedSlack.On("GetUser").Return(mockedSlackUser, nil)
	for _, mu := range mockedUser {
		mockedDB.On("CreateUser", mu).Return(nil)
	}

	h := &router{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}
	err := h.SyncUser()

	assert := assert.New(t)
	mockedGitLab.AssertNumberOfCalls(t, "GetUser", 1)
	mockedSlack.AssertNumberOfCalls(t, "GetUser", 1)
	mockedDB.AssertNumberOfCalls(t, "CreateUser", len(mockedUser))
	assert.Nil(err)
}

func TestDropSomeUser(t *testing.T) {
	mockedDB := &mDB.Store{}
	mockedSlack := &mSlack.Slack{}
	mockedGitLab := &mGitLab.GitLab{}

	var mockedGitLabUser []*gitlab.GitLabUser
	for i := 0; i < 7; i++ {
		fakeEmail := fmt.Sprintf("fake-email-%d@fake.com", i)
		fakeName := fmt.Sprintf("fake-name-%d", i)
		g := &gitlab.GitLabUser{ID: i, Email: fakeEmail, Name: fakeName}
		mockedGitLabUser = append(mockedGitLabUser, g)
	}

	var mockedSlackUser []*slack.SlackUser
	for i := 3; i < 10; i++ {
		fakeEmail := fmt.Sprintf("fake-email-%d@fake.com", i)
		fakeSlackID := fmt.Sprintf("fake-slack-id-%d", i)
		s := &slack.SlackUser{ID: fakeSlackID, Email: fakeEmail}
		mockedSlackUser = append(mockedSlackUser, s)
	}

	var mockedUser []*model.User
	for i := 0; i < 3; i++ {
		fakeEmail := fmt.Sprintf("fake-email-%d@fake.com", i)
		fakeName := fmt.Sprintf("fake-name-%d", i)
		u := &model.User{
			Email:    strings.Split(fakeEmail, "@")[0],
			GitLabID: i,
			Name:     fakeName,
		}
		mockedUser = append(mockedUser, u)
	}
	for i := 3; i < 7; i++ {
		fakeEmail := fmt.Sprintf("fake-email-%d@fake.com", i)
		fakeSlackID := fmt.Sprintf("fake-slack-id-%d", i)
		fakeName := fmt.Sprintf("fake-name-%d", i)
		u := &model.User{
			Email:    strings.Split(fakeEmail, "@")[0],
			SlackID:  fakeSlackID,
			GitLabID: i,
			Name:     fakeName,
		}
		mockedUser = append(mockedUser, u)
	}

	mockedGitLab.On("GetUser").Return(mockedGitLabUser, nil)
	mockedSlack.On("GetUser").Return(mockedSlackUser, nil)
	for _, mu := range mockedUser {
		mockedDB.On("CreateUser", mu).Return(nil)
	}

	h := &router{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}
	err := h.SyncUser()

	assert := assert.New(t)
	mockedGitLab.AssertNumberOfCalls(t, "GetUser", 1)
	mockedSlack.AssertNumberOfCalls(t, "GetUser", 1)
	mockedDB.AssertNumberOfCalls(t, "CreateUser", len(mockedUser))
	assert.Nil(err)
}

func TestGitLabGetUserFail(t *testing.T) {
	mockedGitLab := &mGitLab.GitLab{}

	mockedGitLab.On("GetUser").Return(nil, errors.New("fake-error"))

	h := &router{
		g: mockedGitLab,
	}
	err := h.SyncUser()

	assert := assert.New(t)
	mockedGitLab.AssertNumberOfCalls(t, "GetUser", 1)
	assert.NotNil(err)
	assert.Equal("fake-error", err.Error(), "error message should be equal")
}

func TestSlackGetUserFail(t *testing.T) {
	mockedGitLab := &mGitLab.GitLab{}
	mockedSlack := &mSlack.Slack{}

	// don't care gitlab in this case
	mockedGitLab.On("GetUser").Return(nil, nil)
	mockedSlack.On("GetUser").Return(nil, errors.New("fake-error"))

	h := &router{
		g: mockedGitLab,
		s: mockedSlack,
	}
	err := h.SyncUser()

	assert := assert.New(t)
	mockedGitLab.AssertNumberOfCalls(t, "GetUser", 1)
	mockedSlack.AssertNumberOfCalls(t, "GetUser", 1)
	assert.NotNil(err)
	assert.Equal("fake-error", err.Error(), "error message should be equal")
}

func TestCreateUserFail(t *testing.T) {
	mockedDB := &mDB.Store{}
	mockedSlack := &mSlack.Slack{}
	mockedGitLab := &mGitLab.GitLab{}

	var mockedGitLabUser []*gitlab.GitLabUser
	var mockedSlackUser []*slack.SlackUser
	var mockedUser []*model.User
	for i := 0; i < 10; i++ {
		fakeEmail := fmt.Sprintf("fake-email-%d@fake.com", i)
		fakeName := fmt.Sprintf("fake-name-%d", i)
		fakeSlackID := fmt.Sprintf("fake-slack-id-%d", i)
		g := &gitlab.GitLabUser{ID: i, Email: fakeEmail, Name: fakeName}
		mockedGitLabUser = append(mockedGitLabUser, g)
		s := &slack.SlackUser{ID: fakeSlackID, Email: fakeEmail}
		mockedSlackUser = append(mockedSlackUser, s)
		u := &model.User{
			Email:    strings.Split(fakeEmail, "@")[0],
			SlackID:  s.ID,
			GitLabID: g.ID,
			Name:     g.Name,
		}
		mockedUser = append(mockedUser, u)
	}

	mockedGitLab.On("GetUser").Return(mockedGitLabUser, nil)
	mockedSlack.On("GetUser").Return(mockedSlackUser, nil)
	for i := 0; i < 5; i++ {
		mockedDB.On("CreateUser", mockedUser[i]).Return(errors.New("fake-error"))
	}
	for i := 5; i < 10; i++ {
		mockedDB.On("CreateUser", mockedUser[i]).Return(nil)
	}

	h := &router{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}
	err := h.SyncUser()

	assert := assert.New(t)
	mockedGitLab.AssertNumberOfCalls(t, "GetUser", 1)
	mockedSlack.AssertNumberOfCalls(t, "GetUser", 1)
	mockedDB.AssertNumberOfCalls(t, "CreateUser", len(mockedUser))
	assert.NotNil(err, "error should be nil")
	for i := 0; i < 5; i++ {
		assert.Contains(err.Error(), mockedUser[i].Email)
	}
}
