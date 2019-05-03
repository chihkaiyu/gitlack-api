package webhook

import (
	"fmt"
	"testing"

	"gitlack/model"

	"gitlack/resource/gitlab"
	"gitlack/resource/slack"

	mGitLab "gitlack/resource/gitlab/mocks"
	mSlack "gitlack/resource/slack/mocks"
	mDB "gitlack/store/mocks"
)

const tagPushBodyTemplate = `{
	"checkout_sha": "{{.SHA}}",
	"message": "{{.Message}}",
	"user_id": {{.UserID}},
	"project": {
		"id": {{.ProjectID}},
		"web_url": "http://fake.com/{{.Path}}",
		"path_with_namespace": "{{.Path}}"
    }
}`

func TestTagPushEvent(t *testing.T) {
	fakeData := map[string]interface{}{
		"SHA":       "fake-checkout-sha",
		"Message":   "fake-message",
		"UserID":    1,
		"ProjectID": 999,
		"Path":      "fake/fake-gitlab-project",
	}

	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}
	mockedSlack := &mSlack.Slack{}

	mockedTagList := []*gitlab.Tag{
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc"},
			Name:        "fake-tag-name",
		},
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc2"},
			Name:        "fake-tag-name2",
		},
	}
	mockedAuthor := &model.User{
		Email:   "fake-author@fake.com",
		SlackID: "fake-author-slack-id",
	}
	mockedProject := &model.Project{}

	expected := map[string]string{
		"Author": mockedAuthor.SlackID,
		"Tag":    mockedTagList[0].Name,
		"Path":   fakeData["Path"].(string),
		"Note":   mockedTagList[0].ReleaseInfo.Description,
		"Link":   fmt.Sprintf("http://fake.com/%v/tags/%v", fakeData["Path"].(string), mockedTagList[0].Name),
	}
	slackExpected := renderTemplate(tagPushTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedGitLab.On("GetTagList", fakeData["ProjectID"].(int)).Return(mockedTagList, nil)
	mockedDB.On("GetUserByID", fakeData["UserID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedSlack.On("PostSlackMessage", "general", slackExpected.String(), nilUser, nilAtm).Return(nil, nil)

	w := &hook{
		db: mockedDB,
		g:  mockedGitLab,
		s:  mockedSlack,
	}

	body := renderTemplate(tagPushBodyTemplate, fakeData)
	w.TagPushEvent(body.Bytes())

	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedGitLab.AssertNumberOfCalls(t, "GetTagList", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}

func TestTagPushTagDeleteEvent(t *testing.T) {
	fakeData := map[string]interface{}{
		"SHA":       "",
		"Message":   "fake-message",
		"UserID":    1,
		"ProjectID": 999,
		"Path":      "fake/fake-gitlab-project",
	}

	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}
	mockedSlack := &mSlack.Slack{}

	w := &hook{
		db: mockedDB,
		g:  mockedGitLab,
		s:  mockedSlack,
	}

	body := renderTemplate(tagPushBodyTemplate, fakeData)
	w.TagPushEvent(body.Bytes())

	mockedGitLab.AssertNotCalled(t, "GetTagList")
	mockedDB.AssertNotCalled(t, "GetUserByID")
	mockedDB.AssertNotCalled(t, "GetProjectByID")
	mockedSlack.AssertNotCalled(t, "PostSlackMessage")
}

func TestTagPushOnlyOneTag(t *testing.T) {
	fakeData := map[string]interface{}{
		"SHA":       "fake-checkout-sha",
		"Message":   "fake-message",
		"UserID":    1,
		"ProjectID": 999,
		"Path":      "fake/fake-gitlab-project",
	}

	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}
	mockedSlack := &mSlack.Slack{}

	mockedTagList := []*gitlab.Tag{
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc"},
			Name:        "fake-tag-name",
		},
	}
	mockedAuthor := &model.User{
		Email:   "fake-author@fake.com",
		SlackID: "fake-author-slack-id",
	}
	mockedProject := &model.Project{}

	expected := map[string]string{
		"Author": mockedAuthor.SlackID,
		"Tag":    mockedTagList[0].Name,
		"Path":   fakeData["Path"].(string),
		"Note":   mockedTagList[0].ReleaseInfo.Description,
		"Link":   fmt.Sprintf("http://fake.com/%v/tags/%v", fakeData["Path"].(string), mockedTagList[0].Name),
	}
	slackExpected := renderTemplate(tagPushTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedGitLab.On("GetTagList", fakeData["ProjectID"].(int)).Return(mockedTagList, nil)
	mockedDB.On("GetUserByID", fakeData["UserID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedSlack.On("PostSlackMessage", "general", slackExpected.String(), nilUser, nilAtm).Return(nil, nil)

	w := &hook{
		db: mockedDB,
		g:  mockedGitLab,
		s:  mockedSlack,
	}

	body := renderTemplate(tagPushBodyTemplate, fakeData)
	w.TagPushEvent(body.Bytes())

	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedGitLab.AssertNumberOfCalls(t, "GetTagList", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}

func TestTagPushChannelFromMessage(t *testing.T) {
	fakeData := map[string]interface{}{
		"SHA":       "fake-checkout-sha",
		"UserID":    1,
		"ProjectID": 999,
		"Path":      "fake/fake-gitlab-project",
	}

	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}
	mockedSlack := &mSlack.Slack{}

	mockedTagList := []*gitlab.Tag{
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc"},
			Name:        "fake-tag-name",
		},
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc2"},
			Name:        "fake-tag-name2",
		},
	}
	mockedAuthor := &model.User{
		Email:   "fake-author@fake.com",
		SlackID: "fake-author-slack-id",
	}
	mockedProject := &model.Project{}

	expected := map[string]string{
		"Author": mockedAuthor.SlackID,
		"Tag":    mockedTagList[0].Name,
		"Path":   fakeData["Path"].(string),
		"Note":   mockedTagList[0].ReleaseInfo.Description,
		"Link":   fmt.Sprintf("http://fake.com/%v/tags/%v", fakeData["Path"].(string), mockedTagList[0].Name),
	}
	slackExpected := renderTemplate(tagPushTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedGitLab.On("GetTagList", fakeData["ProjectID"].(int)).Return(mockedTagList, nil)
	mockedDB.On("GetUserByID", fakeData["UserID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)

	input := map[string]string{
		"/gitlack: fake-channel": "fake-channel",
		"/gitlack:fake-channel":  "fake-channel",
		"/gitlack: fake_channel": "fake_channel",
	}
	var w *hook

	for m, c := range input {
		w = &hook{
			db: mockedDB,
			g:  mockedGitLab,
			s:  mockedSlack,
		}

		fakeData["Message"] = fmt.Sprintf("a\\nb\\n%v", m)
		mockedSlack.On("PostSlackMessage", c, slackExpected.String(), nilUser, nilAtm).Return(nil, nil)
		body := renderTemplate(tagPushBodyTemplate, fakeData)
		w.TagPushEvent(body.Bytes())
	}

	mockedGitLab.AssertNumberOfCalls(t, "GetTagList", 1*len(input))
	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1*len(input))
	mockedDB.AssertNotCalled(t, "GetProjectByID")
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1*len(input))
}

func TestTagPushChannelFromProject(t *testing.T) {
	fakeData := map[string]interface{}{
		"SHA":       "fake-checkout-sha",
		"Message":   "fake-message",
		"UserID":    1,
		"ProjectID": 999,
		"Path":      "fake/fake-gitlab-project",
	}

	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}
	mockedSlack := &mSlack.Slack{}

	mockedTagList := []*gitlab.Tag{
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc"},
			Name:        "fake-tag-name",
		},
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc2"},
			Name:        "fake-tag-name2",
		},
	}
	mockedAuthor := &model.User{
		Email:   "fake-author@fake.com",
		SlackID: "fake-author-slack-id",
	}
	mockedProject := &model.Project{DefaultChannel: "fake-project-default-channel"}

	expected := map[string]string{
		"Author": mockedAuthor.SlackID,
		"Tag":    mockedTagList[0].Name,
		"Path":   fakeData["Path"].(string),
		"Note":   mockedTagList[0].ReleaseInfo.Description,
		"Link":   fmt.Sprintf("http://fake.com/%v/tags/%v", fakeData["Path"].(string), mockedTagList[0].Name),
	}
	slackExpected := renderTemplate(tagPushTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedGitLab.On("GetTagList", fakeData["ProjectID"].(int)).Return(mockedTagList, nil)
	mockedDB.On("GetUserByID", fakeData["UserID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedSlack.On("PostSlackMessage", mockedProject.DefaultChannel, slackExpected.String(), nilUser, nilAtm).Return(nil, nil)

	w := &hook{
		db: mockedDB,
		g:  mockedGitLab,
		s:  mockedSlack,
	}

	body := renderTemplate(tagPushBodyTemplate, fakeData)
	w.TagPushEvent(body.Bytes())

	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedGitLab.AssertNumberOfCalls(t, "GetTagList", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}

func TestTagPushChannelFromAuthor(t *testing.T) {
	fakeData := map[string]interface{}{
		"SHA":       "fake-checkout-sha",
		"Message":   "fake-message",
		"UserID":    1,
		"ProjectID": 999,
		"Path":      "fake/fake-gitlab-project",
	}

	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}
	mockedSlack := &mSlack.Slack{}

	mockedTagList := []*gitlab.Tag{
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc"},
			Name:        "fake-tag-name",
		},
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc2"},
			Name:        "fake-tag-name2",
		},
	}
	mockedAuthor := &model.User{
		Email:          "fake-author@fake.com",
		SlackID:        "fake-author-slack-id",
		DefaultChannel: "fake-author-default-channel",
	}
	mockedProject := &model.Project{}

	expected := map[string]string{
		"Author": mockedAuthor.SlackID,
		"Tag":    mockedTagList[0].Name,
		"Path":   fakeData["Path"].(string),
		"Note":   mockedTagList[0].ReleaseInfo.Description,
		"Link":   fmt.Sprintf("http://fake.com/%v/tags/%v", fakeData["Path"].(string), mockedTagList[0].Name),
	}
	slackExpected := renderTemplate(tagPushTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedGitLab.On("GetTagList", fakeData["ProjectID"].(int)).Return(mockedTagList, nil)
	mockedDB.On("GetUserByID", fakeData["UserID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedSlack.On("PostSlackMessage", mockedAuthor.DefaultChannel, slackExpected.String(), nilUser, nilAtm).Return(nil, nil)

	w := &hook{
		db: mockedDB,
		g:  mockedGitLab,
		s:  mockedSlack,
	}

	body := renderTemplate(tagPushBodyTemplate, fakeData)
	w.TagPushEvent(body.Bytes())

	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1)
	mockedDB.AssertNumberOfCalls(t, "GetProjectByID", 1)
	mockedGitLab.AssertNumberOfCalls(t, "GetTagList", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}

func TestTagPushChannelOverwrite(t *testing.T) {
	fakeData := map[string]interface{}{
		"SHA":       "fake-checkout-sha",
		"Message":   "fake-message\\n/gitlack: fake-channel",
		"UserID":    1,
		"ProjectID": 999,
		"Path":      "fake/fake-gitlab-project",
	}

	mockedDB := &mDB.Store{}
	mockedGitLab := &mGitLab.GitLab{}
	mockedSlack := &mSlack.Slack{}

	mockedTagList := []*gitlab.Tag{
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc"},
			Name:        "fake-tag-name",
		},
		&gitlab.Tag{
			ReleaseInfo: gitlab.Release{Description: "fake-desc2"},
			Name:        "fake-tag-name2",
		},
	}
	mockedAuthor := &model.User{
		Email:          "fake-author@fake.com",
		SlackID:        "fake-author-slack-id",
		DefaultChannel: "fake-author-default-channel",
	}
	mockedProject := &model.Project{}

	expected := map[string]string{
		"Author": mockedAuthor.SlackID,
		"Tag":    mockedTagList[0].Name,
		"Path":   fakeData["Path"].(string),
		"Note":   mockedTagList[0].ReleaseInfo.Description,
		"Link":   fmt.Sprintf("http://fake.com/%v/tags/%v", fakeData["Path"].(string), mockedTagList[0].Name),
	}
	slackExpected := renderTemplate(tagPushTemplate, expected)
	var nilUser *model.User
	var nilAtm *slack.Attachment
	mockedGitLab.On("GetTagList", fakeData["ProjectID"].(int)).Return(mockedTagList, nil)
	mockedDB.On("GetUserByID", fakeData["UserID"].(int)).Return(mockedAuthor, nil)
	mockedDB.On("GetProjectByID", fakeData["ProjectID"].(int)).Return(mockedProject, nil)
	mockedSlack.On("PostSlackMessage", "fake-channel", slackExpected.String(), nilUser, nilAtm).Return(nil, nil)

	w := &hook{
		db: mockedDB,
		g:  mockedGitLab,
		s:  mockedSlack,
	}

	body := renderTemplate(tagPushBodyTemplate, fakeData)
	w.TagPushEvent(body.Bytes())

	mockedDB.AssertNumberOfCalls(t, "GetUserByID", 1)
	mockedDB.AssertNotCalled(t, "GetProjectByID")
	mockedGitLab.AssertNumberOfCalls(t, "GetTagList", 1)
	mockedSlack.AssertNumberOfCalls(t, "PostSlackMessage", 1)
}
