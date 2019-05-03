package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlack/resource/slack"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlack/model"
	"gitlack/resource/gitlab"

	mGitLab "gitlack/resource/gitlab/mocks"
	mSlack "gitlack/resource/slack/mocks"
	mDB "gitlack/store/mocks"
)

var mergeRequesBodyTemplate = `
{
	"project": {
		"id": {{.ProjectID}},
		"web_url": "http://fake.com/{{.Path}}",
		"path_with_namespace": "{{.Path}}"
	},
	"object_attributes": {
		"action": "{{.Action}}",
		"assignee_id": {{.AssigneeID}},
		"author_id": {{.AuthorID}},
		"description": "{{.Desc}}",
		"title": "{{.Title}}",
		"source_branch": "{{.Source}}",
		"target_branch": "{{.Target}}",
		"url": "http://fake.com/{{.Path}}/merge_requests/1",
		"iid": {{.ObjectNum}},
		"last_commit": {
			"id": "{{.Sha}}"
		}
	}
}`

func getMRFakeData() map[string]interface{} {
	return map[string]interface{}{
		"ProjectID":  999,
		"Sha":        "fake-commit-sha",
		"Path":       "fake/fake-gitlab-project",
		"Action":     "open",
		"AssigneeID": 1,
		"AuthorID":   2,
		"Desc":       "fake-description",
		"Title":      "fake-title",
		"Source":     "fake-source-branch",
		"Target":     "fake-target-branch",
		"ObjectNum":  1,
	}
}

func renderTemplate(tpl string, data interface{}) *bytes.Buffer {
	t := template.Must(template.New("tpl").Parse(tpl))
	rendered := &bytes.Buffer{}
	t.Execute(rendered, data)
	return rendered
}

func genMRBody(data map[string]interface{}) []byte {
	body := renderTemplate(mergeRequesBodyTemplate, data)
	return body.Bytes()
}

func TestActiveMR(t *testing.T) {
	// prepare fake input
	fakeData := getMRFakeData()
	mockedProject := &model.Project{
		ID:   fakeData["ProjectID"].(int),
		Name: fakeData["Path"].(string),
	}
	mockedAssignee := &model.User{
		SlackID: "fake-assignee-slack-id",
	}
	mockedAuthor := &model.User{
		SlackID: "fake-author-slack-id",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedMR := &model.MergeRequest{
		ProjectID:       fakeData["ProjectID"].(int),
		MergeRequestNum: fakeData["ObjectNum"].(int),
		ThreadTS:        mockedMessageReponse.TS,
		Channel:         mockedMessageReponse.Channel,
	}

	mockedCommit := &gitlab.Commit{
		LastPipeline: gitlab.Pipeline{Status: "success"},
	}

	mockedGitLab := &mGitLab.GitLab{}
	mockedGitLab.On("GetSingleCommit", fakeData["ProjectID"].(int), fakeData["Sha"].(string)).Return(mockedCommit, nil)

	mockedDB := &mDB.Store{}
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("GetUserByID", fakeData["AssigneeID"].(int)).Return(mockedAssignee, nil)
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("CreateMergeRequest", mockedMR).Return(nil)

	// assert Slack text format
	expected := map[string]interface{}{
		"Assignee": mockedAssignee.SlackID,
		"Path":     fakeData["Path"].(string),
		"Author":   mockedAuthor.SlackID,
		"Title":    fakeData["Title"].(string),
		"Source":   fakeData["Source"].(string),
		"Target":   fakeData["Target"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/merge_requests/1", fakeData["Path"].(string)),
		"MRNum":    fakeData["ObjectNum"].(int),
	}
	slackExpected := renderTemplate(mrTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", "general", slackExpected.String(), nilUser, nilAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}

	// mock sleep function
	sleep = func(d time.Duration) {
		// do nothing here
	}

	w.MergeRequestEvent(genMRBody(fakeData))

	// wait for goroutine, work around
	time.Sleep(time.Millisecond * 100)

	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 1)
	mockedGitLab.AssertNumberOfCalls(t, "GetSingleCommit", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)

	// clean up
	sleep = time.Sleep
}

func TestMRSamePerson(t *testing.T) {
	// prepare fake input
	fakeData := getMRFakeData()
	fakeData["AssigneeID"] = fakeData["AuthorID"]

	mockedSlack := &mSlack.Slack{}
	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}
	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}

	w.MergeRequestEvent(genMRBody(fakeData))

	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 0)
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 0)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 0)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 0)
}

func TestMRChannelFromDescription(t *testing.T) {
	// prepare fake input
	fakeData := getMRFakeData()
	mockedProject := &model.Project{
		ID:   fakeData["ProjectID"].(int),
		Name: fakeData["Path"].(string),
	}
	mockedAssignee := &model.User{
		SlackID: "fake-assignee-slack-id",
	}
	mockedAuthor := &model.User{
		SlackID: "fake-author-slack-id",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedMR := &model.MergeRequest{
		ProjectID:       fakeData["ProjectID"].(int),
		MergeRequestNum: fakeData["ObjectNum"].(int),
		ThreadTS:        mockedMessageReponse.TS,
		Channel:         mockedMessageReponse.Channel,
	}
	mockedCommit := &gitlab.Commit{
		LastPipeline: gitlab.Pipeline{Status: "success"},
	}

	mockedGitLab := &mGitLab.GitLab{}
	mockedGitLab.On("GetSingleCommit", fakeData["ProjectID"].(int), fakeData["Sha"].(string)).Return(mockedCommit, nil)

	mockedDB := &mDB.Store{}
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("GetUserByID", fakeData["AssigneeID"].(int)).Return(mockedAssignee, nil)
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("CreateMergeRequest", mockedMR).Return(nil)

	// assert Slack text format
	expected := map[string]interface{}{
		"Assignee": mockedAssignee.SlackID,
		"Path":     fakeData["Path"].(string),
		"Author":   mockedAuthor.SlackID,
		"Title":    fakeData["Title"].(string),
		"Source":   fakeData["Source"].(string),
		"Target":   fakeData["Target"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/merge_requests/1", fakeData["Path"].(string)),
		"MRNum":    fakeData["ObjectNum"].(int),
	}
	slackExpected := renderTemplate(mrTemplate, expected)
	mockedSlack := &mSlack.Slack{}

	var nilUser *model.User
	var nilAtm *slack.Attachment
	input := map[string]string{
		"/gitlack: fake-channel-1": "fake-channel-1",
		"/gitlack:fake-channel-2":  "fake-channel-2",
		"/gitlack: fake_channel-3": "fake_channel-3",
	}

	// mock sleep function
	sleep = func(d time.Duration) {
		// do nothing here
	}

	var w *hook
	for d, c := range input {
		w = &hook{
			db: mockedDB,
			s:  mockedSlack,
			g:  mockedGitLab,
		}

		fakeData["Desc"] = fmt.Sprintf("a\\nb\\n%v", d)
		mockedSlack.On("PostSlackMessage", c, slackExpected.String(), nilUser, nilAtm).Return(mockedMessageReponse, nil)
		w.MergeRequestEvent(genMRBody(fakeData))

		// wait for goroutine, work around
		time.Sleep(time.Millisecond * 100)
	}
	mockedDB.AssertNotCalled(t, "GetProjectByID")
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2*len(input))
	mockedGitLab.AssertNumberOfCalls(t, "GetSingleCommit", 1*len(input))
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1*len(input))

	// clean up
	sleep = time.Sleep
}

func TestMRChannelFromProject(t *testing.T) {
	// prepare fake input
	fakeData := getMRFakeData()
	mockedProject := &model.Project{
		ID:             fakeData["ProjectID"].(int),
		Name:           fakeData["Path"].(string),
		DefaultChannel: "fake-default-project-channel",
	}
	mockedAssignee := &model.User{
		SlackID: "fake-assignee-slack-id",
	}
	mockedAuthor := &model.User{
		SlackID: "fake-author-slack-id",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedMR := &model.MergeRequest{
		ProjectID:       fakeData["ProjectID"].(int),
		MergeRequestNum: fakeData["ObjectNum"].(int),
		ThreadTS:        mockedMessageReponse.TS,
		Channel:         mockedMessageReponse.Channel,
	}
	mockedCommit := &gitlab.Commit{
		LastPipeline: gitlab.Pipeline{Status: "success"},
	}

	mockedGitLab := &mGitLab.GitLab{}
	mockedGitLab.On("GetSingleCommit", fakeData["ProjectID"].(int), fakeData["Sha"].(string)).Return(mockedCommit, nil)

	mockedDB := &mDB.Store{}
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("GetUserByID", fakeData["AssigneeID"].(int)).Return(mockedAssignee, nil)
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("CreateMergeRequest", mockedMR).Return(nil)

	// assert Slack text format
	expected := map[string]interface{}{
		"Assignee": mockedAssignee.SlackID,
		"Path":     fakeData["Path"].(string),
		"Author":   mockedAuthor.SlackID,
		"Title":    fakeData["Title"].(string),
		"Source":   fakeData["Source"].(string),
		"Target":   fakeData["Target"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/merge_requests/1", fakeData["Path"].(string)),
		"MRNum":    fakeData["ObjectNum"].(int),
	}
	slackExpected := renderTemplate(mrTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", mockedProject.DefaultChannel, slackExpected.String(), nilUser, nilAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}

	// mock sleep function
	sleep = func(d time.Duration) {
		// do nothing here
	}

	w.MergeRequestEvent(genMRBody(fakeData))

	// wait for goroutine, work around
	time.Sleep(time.Millisecond * 100)

	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 1)
	mockedGitLab.AssertNumberOfCalls(t, "GetSingleCommit", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)

	// clean up
	sleep = time.Sleep
}

func TestMRChannelFromAssignee(t *testing.T) {
	// prepare fake input
	fakeData := getMRFakeData()
	mockedProject := &model.Project{
		ID:   fakeData["ProjectID"].(int),
		Name: fakeData["Path"].(string),
	}
	mockedAssignee := &model.User{
		SlackID:        "fake-assignee-slack-id",
		DefaultChannel: "fake-default-assignee-channel",
	}
	mockedAuthor := &model.User{
		SlackID: "fake-author-slack-id",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedMR := &model.MergeRequest{
		ProjectID:       fakeData["ProjectID"].(int),
		MergeRequestNum: fakeData["ObjectNum"].(int),
		ThreadTS:        mockedMessageReponse.TS,
		Channel:         mockedMessageReponse.Channel,
	}
	mockedCommit := &gitlab.Commit{
		LastPipeline: gitlab.Pipeline{Status: "success"},
	}

	mockedGitLab := &mGitLab.GitLab{}
	mockedGitLab.On("GetSingleCommit", fakeData["ProjectID"].(int), fakeData["Sha"].(string)).Return(mockedCommit, nil)

	mockedDB := &mDB.Store{}
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("GetUserByID", fakeData["AssigneeID"].(int)).Return(mockedAssignee, nil)
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("CreateMergeRequest", mockedMR).Return(nil)

	// assert Slack text format
	expected := map[string]interface{}{
		"Assignee": mockedAssignee.SlackID,
		"Path":     fakeData["Path"].(string),
		"Author":   mockedAuthor.SlackID,
		"Title":    fakeData["Title"].(string),
		"Source":   fakeData["Source"].(string),
		"Target":   fakeData["Target"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/merge_requests/1", fakeData["Path"].(string)),
		"MRNum":    fakeData["ObjectNum"].(int),
	}
	slackExpected := renderTemplate(mrTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", mockedAssignee.DefaultChannel, slackExpected.String(), nilUser, nilAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}

	// mock sleep function
	sleep = func(d time.Duration) {
		// do nothing here
	}

	w.MergeRequestEvent(genMRBody(fakeData))

	// wait for goroutine, work around
	time.Sleep(time.Millisecond * 100)

	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 1)
	mockedGitLab.AssertNumberOfCalls(t, "GetSingleCommit", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)

	// clean up
	sleep = time.Sleep
}

func TestMRChannelOverwrite(t *testing.T) {
	// prepare fake input
	fakeData := getMRFakeData()
	fakeData["Desc"] = fmt.Sprintf("%v%v", fakeData["Desc"].(string), "\\n/gitlack: fake-description-channel")

	mockedProject := &model.Project{
		ID:             fakeData["ProjectID"].(int),
		Name:           fakeData["Path"].(string),
		DefaultChannel: "fake-default-project-channel",
	}
	mockedAssignee := &model.User{
		SlackID:        "fake-assignee-slack-id",
		DefaultChannel: "fake-default-assignee-channel",
	}
	mockedAuthor := &model.User{
		SlackID: "fake-author-slack-id",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedMR := &model.MergeRequest{
		ProjectID:       fakeData["ProjectID"].(int),
		MergeRequestNum: fakeData["ObjectNum"].(int),
		ThreadTS:        mockedMessageReponse.TS,
		Channel:         mockedMessageReponse.Channel,
	}
	mockedCommit := &gitlab.Commit{
		LastPipeline: gitlab.Pipeline{Status: "success"},
	}

	mockedGitLab := &mGitLab.GitLab{}
	mockedGitLab.On("GetSingleCommit", fakeData["ProjectID"].(int), fakeData["Sha"].(string)).Return(mockedCommit, nil)

	mockedDB := &mDB.Store{}
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("GetUserByID", fakeData["AssigneeID"].(int)).Return(mockedAssignee, nil)
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("CreateMergeRequest", mockedMR).Return(nil)

	// assert Slack text format
	expected := map[string]interface{}{
		"Assignee": mockedAssignee.SlackID,
		"Path":     fakeData["Path"].(string),
		"Author":   mockedAuthor.SlackID,
		"Title":    fakeData["Title"].(string),
		"Source":   fakeData["Source"].(string),
		"Target":   fakeData["Target"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/merge_requests/1", fakeData["Path"].(string)),
		"MRNum":    fakeData["ObjectNum"].(int),
	}
	slackExpected := renderTemplate(mrTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", "fake-description-channel", slackExpected.String(), nilUser, nilAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}

	// mock sleep function
	sleep = func(d time.Duration) {
		// do nothing here
	}

	w.MergeRequestEvent(genMRBody(fakeData))

	// wait for goroutine, work around
	time.Sleep(time.Millisecond * 100)

	mockedDB.AssertNotCalled(t, "GetProjectByID")
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 1)
	mockedGitLab.AssertNumberOfCalls(t, "GetSingleCommit", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)

	// clean up
	sleep = time.Sleep
}

func TestDeactiveMR(t *testing.T) {
	// prepare fake input
	fakeData := getMRFakeData()

	input := map[string]string{
		"merge": "This merge request has been merged.",
		"close": "This merge request has been closed.",
	}

	mockedMR := &model.MergeRequest{
		ProjectID:       fakeData["ProjectID"].(int),
		MergeRequestNum: fakeData["ObjectNum"].(int),
		ThreadTS:        "1234567890.123456",
		Channel:         "fake-channel",
	}

	mockedDB := &mDB.Store{}
	mockedDB.On("GetMergeRequest", fakeData["ProjectID"].(int), fakeData["ObjectNum"].(int)).Return(mockedMR, nil)
	mockedSlack := &mSlack.Slack{}

	var nilUser *model.User
	var nilAtm *slack.Attachment
	for a, e := range input {
		fakeData["Action"] = a
		w := &hook{
			db: mockedDB,
			s:  mockedSlack,
		}
		mockedSlack.On("PostSlackMessage", mockedMR.Channel, e, nilUser, nilAtm, mockedMR.ThreadTS).Return(nil, nil)
		w.MergeRequestEvent(genMRBody(fakeData))
	}
}

func TestTrackPipelineStatusSuccess(t *testing.T) {
	fakeData := getMRFakeData()

	mockedCommit := &gitlab.Commit{
		LastPipeline: gitlab.Pipeline{Status: "success"},
	}

	mockedGitLab := &mGitLab.GitLab{}
	mockedGitLab.On("GetSingleCommit", fakeData["ProjectID"].(int), fakeData["Sha"].(string)).Return(mockedCommit, nil)

	mockedDB := &mDB.Store{}
	mockedSlack := &mSlack.Slack{}

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}

	assert := assert.New(t)
	var mr MergeRequestEvent
	err := json.Unmarshal(genMRBody(fakeData), &mr)
	assert.Nil(err)

	// mock sleep function
	sleep = func(d time.Duration) {
		// do nothing here
	}

	trackPipelineStatus(mr, w)

	mockedGitLab.AssertNumberOfCalls(t, "GetSingleCommit", 1)
	mockedDB.AssertNotCalled(t, "GetMergeRequest")
	mockedSlack.AssertNotCalled(t, "PostSlackMessage")

	// clean up
	sleep = time.Sleep
}

func TestTrackPipelineStatusFail(t *testing.T) {
	fakeData := getMRFakeData()

	mockedCommit := &gitlab.Commit{
		LastPipeline: gitlab.Pipeline{
			ID:     fakeData["ProjectID"].(int),
			Status: "failed",
			WebURL: fakeData["Path"].(string),
		},
	}
	mockedMR := &model.MergeRequest{
		ProjectID:       fakeData["ProjectID"].(int),
		MergeRequestNum: fakeData["ObjectNum"].(int),
		ThreadTS:        "1234567890.123456",
		Channel:         "fake-channel",
	}

	mockedGitLab := &mGitLab.GitLab{}
	mockedGitLab.On("GetSingleCommit", fakeData["ProjectID"].(int), fakeData["Sha"].(string)).Return(mockedCommit, nil)

	mockedDB := &mDB.Store{}
	mockedDB.On("GetMergeRequest", fakeData["ProjectID"].(int), fakeData["ObjectNum"].(int)).Return(mockedMR, nil)

	slackExpected := fmt.Sprintf("Pipeline <%v|#%v> failed!", fakeData["Path"].(string), fakeData["ProjectID"].(int))
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", mockedMR.Channel, slackExpected, nilUser, nilAtm, mockedMR.ThreadTS).Return(nil, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}

	// mock sleep function
	sleep = func(d time.Duration) {
		// do nothing here
	}

	assert := assert.New(t)
	var mr MergeRequestEvent
	err := json.Unmarshal(genMRBody(fakeData), &mr)
	assert.Nil(err)

	trackPipelineStatus(mr, w)

	mockedGitLab.AssertNumberOfCalls(t, "GetSingleCommit", 1)
	mockedDB.AssertNotCalled(t, "GetMergeRequest")
	mockedSlack.AssertNotCalled(t, "PostSlackMessage")

	// clean up
	sleep = time.Sleep
}

func TestTrackPipelineStatusTimeout(t *testing.T) {
	fakeData := getMRFakeData()

	mockedCommit := &gitlab.Commit{
		LastPipeline: gitlab.Pipeline{Status: "running"},
	}

	mockedGitLab := &mGitLab.GitLab{}
	mockedGitLab.On("GetSingleCommit", fakeData["ProjectID"].(int), fakeData["Sha"].(string)).Return(mockedCommit, nil)

	mockedDB := &mDB.Store{}
	mockedSlack := &mSlack.Slack{}

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
		g:  mockedGitLab,
	}

	assert := assert.New(t)
	var mr MergeRequestEvent
	err := json.Unmarshal(genMRBody(fakeData), &mr)
	assert.Nil(err)

	// mock sleep function
	sleep = func(d time.Duration) {
		// do nothing here
	}

	trackPipelineStatus(mr, w)

	mockedGitLab.AssertNumberOfCalls(t, "GetSingleCommit", 120)
	mockedDB.AssertNotCalled(t, "GetMergeRequest")
	mockedSlack.AssertNotCalled(t, "PostSlackMessage")

	// clean up
	sleep = time.Sleep
}
