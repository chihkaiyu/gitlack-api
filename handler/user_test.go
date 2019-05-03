package handler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncUserWithFiveUsers(t *testing.T) {
	// arrange
	stubGitLab := getStubGetUserGitLab(getGitLabUser(0, 5))
	stubSlack := getStubGetUserSlack(getSlackuser(0, 5))
	stubDB := getStubDB("CreateUser", nil)
	router := getRouter(stubDB, stubSlack, stubGitLab)

	// act
	router.SyncUser()

	// assert
	stubDB.AssertNumberOfCalls(t, "CreateUser", 5)
}

func TestSyncUserWithNoError(t *testing.T) {
	// arrange
	stubGitLab := getStubGetUserGitLab(getGitLabUser(0, 5))
	stubSlack := getStubGetUserSlack(getSlackuser(0, 5))
	stubDB := getStubDB("CreateUser", nil)
	router := getRouter(stubDB, stubSlack, stubGitLab)

	// act
	err := router.SyncUser()

	// assert
	assert.NoError(t, err, "Should not have error")
}

func TestSyncUserWithMoreSlackuser(t *testing.T) {
	// arrange
	stubGitLab := getStubGetUserGitLab(getGitLabUser(0, 5))
	stubSlack := getStubGetUserSlack(getSlackuser(0, 10))
	stubDB := getStubDB("CreateUser", nil)
	router := getRouter(stubDB, stubSlack, stubGitLab)

	// act
	router.SyncUser()

	// assert
	stubDB.AssertNumberOfCalls(t, "CreateUser", 5)
}

func TestSyncUserWithMoreGitLabuser(t *testing.T) {
	// arrange
	stubGitLab := getStubGetUserGitLab(getGitLabUser(0, 10))
	stubSlack := getStubGetUserSlack(getSlackuser(0, 5))
	stubDB := getStubDB("CreateUser", nil)
	router := getRouter(stubDB, stubSlack, stubGitLab)

	// act
	router.SyncUser()

	// assert
	stubDB.AssertNumberOfCalls(t, "CreateUser", 10)
}

func TestSyncUserGitLabGetUserCalledOnce(t *testing.T) {
	// arrange
	stubGitLab := getStubGetUserGitLab(getGitLabUser(0, 5))
	stubSlack := getStubGetUserSlack(getSlackuser(0, 5))
	stubDB := getStubDB("CreateUser", nil)
	router := getRouter(stubDB, stubSlack, stubGitLab)

	// act
	router.SyncUser()

	// assert
	stubGitLab.AssertNumberOfCalls(t, "GetUser", 1)
}

func TestSyncUserSlackGetUserCalledOnce(t *testing.T) {
	// arrange
	stubGitLab := getStubGetUserGitLab(getGitLabUser(0, 5))
	stubSlack := getStubGetUserSlack(getSlackuser(0, 5))
	stubDB := getStubDB("CreateUser", nil)
	router := getRouter(stubDB, stubSlack, stubGitLab)

	// act
	router.SyncUser()

	// assert
	stubSlack.AssertNumberOfCalls(t, "GetUser", 1)
}

func TestSyncUserWithCreateUserFail(t *testing.T) {
	// arrange
	stubGitLab := getStubGetUserGitLab(getGitLabUser(0, 5))
	stubSlack := getStubGetUserSlack(getSlackuser(0, 5))
	stubDB := getStubDB("CreateUser", fmt.Errorf("fake-error"))
	router := getRouter(stubDB, stubSlack, stubGitLab)

	// act
	err := router.SyncUser()

	// assert
	assert.Error(t, err, "Error should not be nil")
}
