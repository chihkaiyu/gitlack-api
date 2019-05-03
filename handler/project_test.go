package handler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncProjectWithFiveProjects(t *testing.T) {
	// arrange
	stubGitLab := getStubGetProjectGitLab(getProjects(5))
	stubDB := getStubDB("CreateProject", nil)
	router := getRouter(stubDB, nil, stubGitLab)

	// act
	router.SyncProject()

	// assert
	stubDB.AssertNumberOfCalls(t, "CreateProject", 5)
}

func TestSyncProjectWithNoError(t *testing.T) {
	// arrange
	stubGitLab := getStubGetProjectGitLab(getProjects(5))
	stubDB := getStubDB("CreateProject", nil)
	router := getRouter(stubDB, nil, stubGitLab)

	// act
	err := router.SyncProject()

	// assert
	assert.NoError(t, err, "Should not have error")
}

func TestSyncProjectWithCreateProjectFail(t *testing.T) {
	// arrange
	stubGitLab := getStubGetProjectGitLab(getProjects(5))
	stubDB := getStubDB("CreateProject", fmt.Errorf("fake-error"))
	router := getRouter(stubDB, nil, stubGitLab)

	// act
	err := router.SyncProject()

	// assert
	assert.Error(t, err, "Error should not be nil")
}
