package webhook

import (
	"bytes"
	"encoding/json"
	"gitlack/model"
	"gitlack/resource/slack"
	"regexp"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

const issueTemplate = "<@{{.Author}}> has opened <{{.Link}}|{{.Path}}#{{.IssueNum}}>"

// IssuesEvent represents the data structure of issues events
type IssuesEvent struct {
	ObjAttr     ObjectAttributes `json:"object_attributes"`
	ProjectInfo Project          `json:"project"`
}

func (h *hook) IssuesEvent(b []byte) {
	var issue IssuesEvent
	err := json.Unmarshal(b, &issue)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	if issue.ObjAttr.Action == "open" || issue.ObjAttr.Action == "reopen" {
		activeIssue(issue, h)
	} else if issue.ObjAttr.Action == "close" {
		deactiveIssue(issue, h)
	} else {
		logrus.Infoln("action is NOT one of open, reopen or close")
		return
	}
}

func activeIssue(issue IssuesEvent, h *hook) {
	author, err := h.db.GetUserByID(issue.ObjAttr.AuthorID)
	if err != nil {
		return
	}

	// get target Slack channel, priority: description > project > user > #general
	// get from issue description
	var channel string
	re, err := regexp.Compile("/gitlack:\\s?\\S+")
	if err != nil {
		logrus.Errorln(err)
		return
	}
	parsed := re.FindString(issue.ObjAttr.Description)
	if parsed != "" {
		channel = strings.TrimSpace(strings.Replace(parsed, "/gitlack:", "", 1))
	}

	// get from project default channel
	if channel == "" {
		proejct, err := h.db.GetProjectByID(issue.ProjectInfo.ID)
		if err != nil {
			return
		}
		if proejct.DefaultChannel != "" {
			channel = proejct.DefaultChannel
		}

	}

	// get from assignee default channel
	if channel == "" {
		if author.DefaultChannel != "" {
			channel = author.DefaultChannel
		}
	}

	if channel == "" {
		channel = "general"
	}

	// prepare Slack text
	attachment := &slack.Attachment{
		Color: slack.AttachmentColor,
		Title: issue.ObjAttr.Title,
		Text:  issue.ObjAttr.Description,
	}
	data := map[string]interface{}{
		"Author":   author.SlackID,
		"Path":     issue.ProjectInfo.PathWithNamespace,
		"Link":     issue.ObjAttr.ObjectURL,
		"IssueNum": issue.ObjAttr.ObjectNum,
	}
	t, err := template.New("slack").Parse(issueTemplate)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	slackText := &bytes.Buffer{}
	err = t.Execute(slackText, data)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	smr, err := h.s.PostSlackMessage(channel, slackText.String(), author, attachment)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	// insert new issue
	newIssue := &model.Issue{
		ProjectID: issue.ProjectInfo.ID,
		IssueNum:  issue.ObjAttr.ObjectNum,
		ThreadTS:  smr.TS,
		Channel:   smr.Channel,
	}
	err = h.db.CreateIssue(newIssue)
	if err != nil {
		return
	}
}

func deactiveIssue(issue IssuesEvent, h *hook) {
	issueThread, err := h.db.GetIssue(issue.ProjectInfo.ID, issue.ObjAttr.ObjectNum)
	if err != nil {
		return
	}
	slackText := "This issue has been closed."
	h.s.PostSlackMessage(issueThread.Channel, slackText, nil, nil, issueThread.ThreadTS)
}
