package store

import (
	"gitlack/model"
)

// Store defines the interface that storage needs
type Store interface {
	GetProjectByPath(string) (*model.Project, error)
	GetProjectByID(int) (*model.Project, error)
	GetUserByEmail(string) (*model.User, error)
	GetUserByID(int) (*model.User, error)
	GetMergeRequest(int, int) (*model.MergeRequest, error)
	GetIssue(int, int) (*model.Issue, error)

	UpdateUserDefaultChannel(string, string) error
	UpdateProjectDefaultChannel(string, string) error

	CreateUser(*model.User) error
	CreateProject(*model.Project) error
	CreateMergeRequest(*model.MergeRequest) error
	CreateIssue(*model.Issue) error
}
