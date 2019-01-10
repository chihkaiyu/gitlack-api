package webhook

import (
	"bytes"
	"fmt"
	"gitlack/resource/slack"
	"testing"
	"text/template"

	"gitlack/model"

	mSlack "gitlack/resource/slack/mocks"
	mDB "gitlack/store/mocks"
)

const mergeRequesBodyTemplate = `
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
		"iid": {{.ObjectNum}}
	}
}`

func getMRFakeData() map[string]interface{} {
	return map[string]interface{}{
		"ProjectID":  999,
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
	var nilAtm *slack.Attachment
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", "general", slackExpected.String(), nilAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
	}

	w.MergeRequestEvent(genMRBody(fakeData))

	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}

func TestMRNotAcceptableAction(t *testing.T) {
	// prepare fake input
	fakeData := getMRFakeData()
	fakeData["Action"] = "update"

	mockedSlack := &mSlack.Slack{}
	mockedDB := &mDB.Store{}
	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
	}

	w.MergeRequestEvent(genMRBody(fakeData))

	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 0)
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 0)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 0)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 0)
}

func TestMRSamePerson(t *testing.T) {
	// prepare fake input
	fakeData := getMRFakeData()
	fakeData["Action"] = "open"
	fakeData["AssigneeID"] = fakeData["AuthorID"]

	mockedSlack := &mSlack.Slack{}
	mockedDB := &mDB.Store{}
	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
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

	var nilAtm *slack.Attachment
	input := map[string]string{
		"/gitlack: fake-channel-1": "fake-channel-1",
		"/gitlack:fake-channel-2":  "fake-channel-2",
		"/gitlack: fake_channel-3": "fake_channel-3",
	}
	var w *hook
	for d, c := range input {
		w = &hook{
			db: mockedDB,
			s:  mockedSlack,
		}

		fakeData["Desc"] = fmt.Sprintf("a\\nb\\n%v", d)
		mockedSlack.On("PostSlackMessage", c, slackExpected.String(), nilAtm).Return(mockedMessageReponse, nil)
		w.MergeRequestEvent(genMRBody(fakeData))
	}
	mockedDB.AssertNotCalled(t, "GetProjectByID")
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2*len(input))
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1*len(input))

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
	var nilAtm *slack.Attachment
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", mockedProject.DefaultChannel, slackExpected.String(), nilAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
	}

	w.MergeRequestEvent(genMRBody(fakeData))

	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
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
	var nilAtm *slack.Attachment
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", mockedAssignee.DefaultChannel, slackExpected.String(), nilAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
	}

	w.MergeRequestEvent(genMRBody(fakeData))

	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
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
	var nilAtm *slack.Attachment
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", "fake-description-channel", slackExpected.String(), nilAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
	}
	w.MergeRequestEvent(genMRBody(fakeData))

	mockedDB.AssertNotCalled(t, "GetProjectByID")
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 2)
	mockedDB.AssertNumberOfCalls(t, "CreateMergeRequest", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
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

	var nilAtm *slack.Attachment
	for a, e := range input {
		fakeData["Action"] = a
		w := &hook{
			db: mockedDB,
			s:  mockedSlack,
		}
		mockedSlack.On("PostSlackMessage", mockedMR.Channel, e, nilAtm, mockedMR.ThreadTS).Return(nil, nil)
		w.MergeRequestEvent(genMRBody(fakeData))
	}
}
