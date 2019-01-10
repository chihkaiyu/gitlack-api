package webhook

import (
	"fmt"
	"gitlack/model"
	"gitlack/resource/slack"
	mSlack "gitlack/resource/slack/mocks"
	mDB "gitlack/store/mocks"
	"testing"
)

const issuesBodyTemplate = `
{
	"project": {
		"id": {{.ProjectID}},
		"web_url": "http://fake.com/{{.Path}}",
		"path_with_namespace": "{{.Path}}"
	},
	"object_attributes": {
		"action": "{{.Action}}",
		"author_id": {{.AuthorID}},
		"description": "{{.Desc}}",
		"title": "{{.Title}}",
		"url": "http://fake.com/{{.Path}}/issues/1",
		"iid": {{.ObjectNum}}
	}
}`

func getIssuesFakeData() map[string]interface{} {
	return map[string]interface{}{
		"ProjectID": 999,
		"Path":      "fake/fake-gitlab-project",
		"Action":    "open",
		"AuthorID":  2,
		"Desc":      "fake-description",
		"Title":     "fake-title",
		"ObjectNum": 1,
	}
}

func genIssuesBody(data map[string]interface{}) []byte {
	body := renderTemplate(issuesBodyTemplate, data)
	return body.Bytes()
}

func TestIssuesEvent(t *testing.T) {
	// prepare fake input
	fakeData := getIssuesFakeData()
	mockedProject := &model.Project{
		ID:   fakeData["ProjectID"].(int),
		Name: fakeData["Path"].(string),
	}
	mockedAuthor := &model.User{
		SlackID: "fake-author-slack-id",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedIssue := &model.Issue{
		ProjectID: fakeData["ProjectID"].(int),
		IssueNum:  fakeData["ObjectNum"].(int),
		ThreadTS:  mockedMessageReponse.TS,
		Channel:   mockedMessageReponse.Channel,
	}

	mockedDB := &mDB.Store{}
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("CreateIssue", mockedIssue).Return(nil)

	// assert Slack text format
	expectedAtm := &slack.Attachment{
		Color: slack.AttachmentColor,
		Title: fakeData["Title"].(string),
		Text:  fakeData["Desc"].(string),
	}
	expected := map[string]interface{}{
		"Author":   mockedAuthor.SlackID,
		"Path":     fakeData["Path"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/issues/1", fakeData["Path"].(string)),
		"IssueNum": fakeData["ObjectNum"].(int),
	}
	slackExpected := renderTemplate(issueTemplate, expected)
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", "general", slackExpected.String(), expectedAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
	}

	w.IssuesEvent(genIssuesBody(fakeData))

	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedDB.AssertNumberOfCalls(t, "CreateIssue", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}

func TestIssuesChannelFromDescription(t *testing.T) {
	// prepare fake input
	fakeData := getIssuesFakeData()
	mockedProject := &model.Project{
		ID:   fakeData["ProjectID"].(int),
		Name: fakeData["Path"].(string),
	}
	mockedAuthor := &model.User{
		SlackID: "fake-author-slack-id",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedIssue := &model.Issue{
		ProjectID: fakeData["ProjectID"].(int),
		IssueNum:  fakeData["ObjectNum"].(int),
		ThreadTS:  mockedMessageReponse.TS,
		Channel:   mockedMessageReponse.Channel,
	}

	mockedDB := &mDB.Store{}
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("CreateIssue", mockedIssue).Return(nil)

	// assert Slack text format
	expected := map[string]interface{}{
		"Author":   mockedAuthor.SlackID,
		"Path":     fakeData["Path"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/issues/1", fakeData["Path"].(string)),
		"IssueNum": fakeData["ObjectNum"].(int),
	}
	expectedAtm := &slack.Attachment{
		Color: slack.AttachmentColor,
		Title: fakeData["Title"].(string),
		Text:  fakeData["Desc"].(string),
	}

	mockedSlack := &mSlack.Slack{}

	input := map[string]string{
		"/gitlack: fake-channel-1": "fake-channel-1",
		"/gitlack:fake-channel-2":  "fake-channel-2",
		"/gitlack: fake_channel-3": "fake_channel-3",
	}

	var w *hook
	for d, c := range input {
		w = &hook{db: mockedDB, s: mockedSlack}

		// should use `\\` as escape in JSON
		fakeData["Desc"] = fmt.Sprintf("a\\nb\\n%v", d)
		expectedAtm.Text = fmt.Sprintf("a\nb\n%v", d)
		slackExpected := renderTemplate(issueTemplate, expected)
		mockedSlack.On("PostSlackMessage", c, slackExpected.String(), expectedAtm).Return(mockedMessageReponse, nil)
		w.IssuesEvent(genIssuesBody(fakeData))
	}

	mockedDB.AssertNotCalled(t, "GetProjectByID")
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1*len(input))
	mockedDB.AssertNumberOfCalls(t, "CreateIssue", 1*len(input))
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1*len(input))
}

func TestIssuesChannelFromProject(t *testing.T) {
	// prepare fake input
	fakeData := getIssuesFakeData()
	mockedProject := &model.Project{
		ID:             fakeData["ProjectID"].(int),
		Name:           fakeData["Path"].(string),
		DefaultChannel: "fake-default-project-channel",
	}
	mockedAuthor := &model.User{
		SlackID: "fake-author-slack-id",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedIssue := &model.Issue{
		ProjectID: fakeData["ProjectID"].(int),
		IssueNum:  fakeData["ObjectNum"].(int),
		ThreadTS:  mockedMessageReponse.TS,
		Channel:   mockedMessageReponse.Channel,
	}

	mockedDB := &mDB.Store{}
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("CreateIssue", mockedIssue).Return(nil)

	// assert Slack text format
	expectedAtm := &slack.Attachment{
		Color: slack.AttachmentColor,
		Title: fakeData["Title"].(string),
		Text:  fakeData["Desc"].(string),
	}
	expected := map[string]interface{}{
		"Author":   mockedAuthor.SlackID,
		"Path":     fakeData["Path"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/issues/1", fakeData["Path"].(string)),
		"IssueNum": fakeData["ObjectNum"].(int),
	}
	slackExpected := renderTemplate(issueTemplate, expected)
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", mockedProject.DefaultChannel, slackExpected.String(), expectedAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
	}

	w.IssuesEvent(genIssuesBody(fakeData))

	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedDB.AssertNumberOfCalls(t, "CreateIssue", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}

func TestIssuesChannelFromAuthor(t *testing.T) {
	// prepare fake input
	fakeData := getIssuesFakeData()
	mockedProject := &model.Project{
		ID:   fakeData["ProjectID"].(int),
		Name: fakeData["Path"].(string),
	}
	mockedAuthor := &model.User{
		SlackID:        "fake-author-slack-id",
		DefaultChannel: "fake-default-author-channel",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedIssue := &model.Issue{
		ProjectID: fakeData["ProjectID"].(int),
		IssueNum:  fakeData["ObjectNum"].(int),
		ThreadTS:  mockedMessageReponse.TS,
		Channel:   mockedMessageReponse.Channel,
	}

	mockedDB := &mDB.Store{}
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("CreateIssue", mockedIssue).Return(nil)

	// assert Slack text format
	expectedAtm := &slack.Attachment{
		Color: slack.AttachmentColor,
		Title: fakeData["Title"].(string),
		Text:  fakeData["Desc"].(string),
	}
	expected := map[string]interface{}{
		"Author":   mockedAuthor.SlackID,
		"Path":     fakeData["Path"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/issues/1", fakeData["Path"].(string)),
		"IssueNum": fakeData["ObjectNum"].(int),
	}
	slackExpected := renderTemplate(issueTemplate, expected)
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", mockedAuthor.DefaultChannel, slackExpected.String(), expectedAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
	}

	w.IssuesEvent(genIssuesBody(fakeData))

	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedDB.AssertNumberOfCalls(t, "CreateIssue", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}

func TestIssuesChannelOverwrite(t *testing.T) {
	// prepare fake input
	fakeData := getIssuesFakeData()
	expectedDesc := fmt.Sprintf("%v%v", fakeData["Desc"].(string), "\n/gitlack: fake-description-channel")
	// use `\\` as escape in JSON
	fakeData["Desc"] = fmt.Sprintf("%v%v", fakeData["Desc"].(string), "\\n/gitlack: fake-description-channel")

	mockedProject := &model.Project{
		ID:             fakeData["ProjectID"].(int),
		Name:           fakeData["Path"].(string),
		DefaultChannel: "fake-default-project-channel",
	}
	mockedAuthor := &model.User{
		SlackID:        "fake-author-slack-id",
		DefaultChannel: "fake-default-author-channel",
	}
	mockedMessageReponse := &slack.MessageResponse{
		OK:      true,
		Channel: "fake-target-channel",
		TS:      "1234567890.123456",
	}
	mockedIssue := &model.Issue{
		ProjectID: fakeData["ProjectID"].(int),
		IssueNum:  fakeData["ObjectNum"].(int),
		ThreadTS:  mockedMessageReponse.TS,
		Channel:   mockedMessageReponse.Channel,
	}

	mockedDB := &mDB.Store{}
	mockedDB.On("GetUserByID", fakeData["AuthorID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedDB.On("CreateIssue", mockedIssue).Return(nil)

	// assert Slack text format
	expectedAtm := &slack.Attachment{
		Color: slack.AttachmentColor,
		Title: fakeData["Title"].(string),
		Text:  expectedDesc,
	}
	expected := map[string]interface{}{
		"Author":   mockedAuthor.SlackID,
		"Path":     fakeData["Path"].(string),
		"Link":     fmt.Sprintf("http://fake.com/%v/issues/1", fakeData["Path"].(string)),
		"IssueNum": fakeData["ObjectNum"].(int),
	}
	slackExpected := renderTemplate(issueTemplate, expected)
	mockedSlack := &mSlack.Slack{}
	mockedSlack.On("PostSlackMessage", "fake-description-channel", slackExpected.String(), expectedAtm).Return(mockedMessageReponse, nil)

	w := &hook{
		db: mockedDB,
		s:  mockedSlack,
	}

	w.IssuesEvent(genIssuesBody(fakeData))

	mockedDB.AssertNotCalled(t, "GetProjectByID")
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1)
	mockedDB.AssertNumberOfCalls(t, "CreateIssue", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}

func TestDeactiveIssues(t *testing.T) {
	// prepare fake input
	fakeData := getIssuesFakeData()
	fakeData["Action"] = "close"

	mockedIssue := &model.Issue{
		ProjectID: fakeData["ProjectID"].(int),
		IssueNum:  fakeData["ObjectNum"].(int),
		ThreadTS:  "1234567890.123456",
		Channel:   "fake-channel",
	}

	mockedDB := &mDB.Store{}
	mockedDB.On("GetIssue", fakeData["ProjectID"].(int), fakeData["ObjectNum"].(int)).Return(mockedIssue, nil)
	mockedSlack := &mSlack.Slack{}

	w := &hook{db: mockedDB, s: mockedSlack}
	var nilAtm *slack.Attachment
	mockedSlack.On("PostSlackMessage", mockedIssue.Channel, "This issue has been closed.", nilAtm, mockedIssue.ThreadTS).Return(nil, nil)
	w.IssuesEvent(genIssuesBody(fakeData))
}
