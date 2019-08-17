package model

// Project is the model of GitLab project
type Project struct {
	ID             int    `db:"id"`
	Name           string `db:"name"`
	DefaultChannel string `db:"default_channel"`
}

// User is the model of user
type User struct {
	Email          string `db:"email"`
	SlackID        string `db:"slack_id"`
	GitLabID       int    `db:"gitlab_id"`
	Name           string `db:"name"`
	AvatarURL      string `db:"avatar_url"`
	DefaultChannel string `db:"default_channel"`
}

// MergeRequest is the model of GitLab merge request
type MergeRequest struct {
	id              int    `db:"id"`
	ProjectID       int    `db:"project_id"`
	MergeRequestNum int    `db:"mr_num"`
	ThreadTS        string `db:"thread_ts"`
	Channel         string `db:"channel"`
}

// Issue is the model of GitLab issue
type Issue struct {
	id        int    `db:"id"`
	ProjectID int    `db:"project_id"`
	IssueNum  int    `db:"issue_num"`
	ThreadTS  string `db:"thread_ts"`
	Channel   string `db:"channel"`
}
