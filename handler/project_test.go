package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlack/model"
	mGitLab "gitlack/resource/gitlab/mocks"
	mDB "gitlack/store/mocks"
)

func TestSyncProject(t *testing.T) {
	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}

	var mockedProject []*model.Project
	for i := 0; i < 10; i++ {
		tmp := &model.Project{
			ID:   i,
			Name: fmt.Sprintf("fake-project-%d", i),
		}
		mockedProject = append(mockedProject, tmp)
	}

	mockedGitLab.On("GetProject").Return(mockedProject, nil)
	for _, mp := range mockedProject {
		mockedDB.On("CreateProject", mp).Return(nil)
	}

	h := &router{
		db: mockedDB,
		g:  mockedGitLab,
	}
	err := h.SyncProject()

	assert := assert.New(t)
	mockedGitLab.AssertNumberOfCalls(t, "GetProject", 1)
	mockedDB.AssertNumberOfCalls(t, "CreateProject", len(mockedProject))
	assert.Nil(err)
}

func TestCreateProjectFail(t *testing.T) {
	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}

	var mockedProject []*model.Project
	for i := 0; i < 10; i++ {
		tmp := &model.Project{
			ID:   i,
			Name: fmt.Sprintf("fake-project-%d", i),
		}
		mockedProject = append(mockedProject, tmp)
	}

	mockedGitLab.On("GetProject").Return(mockedProject, nil)
	var mockedErrorMsg []string
	for i := 0; i < 5; i++ {
		mockedErrorMsg = append(mockedErrorMsg, fmt.Sprintf("fake-project-%d", i))
		mockedDB.On("CreateProject", mockedProject[i]).Return(errors.New("fake-error"))
	}
	for i := 5; i < 10; i++ {
		mockedDB.On("CreateProject", mockedProject[i]).Return(nil)
	}

	h := &router{
		db: mockedDB,
		g:  mockedGitLab,
	}
	err := h.SyncProject()

	expected, _ := json.Marshal(mockedErrorMsg)
	assert := assert.New(t)
	mockedGitLab.AssertNumberOfCalls(t, "GetProject", 1)
	mockedDB.AssertNumberOfCalls(t, "CreateProject", len(mockedProject))
	assert.NotNil(err, "error should be nil")
	assert.Equal(string(expected), err.Error(), "error message should be equal")
}
