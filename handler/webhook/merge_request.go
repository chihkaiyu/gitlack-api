package webhook

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"text/template"

	"gitlack/model"

	"github.com/sirupsen/logrus"
)

const mrTemplate = "<@{{.Assignee}}> you are assigned to review <{{.Link}}|{{.Path}}!{{.MRNum}}> by <@{{.Author}}>\n" +
	"Title: {{.Title}}\n" +
	"Action: request to merge `{{.Source}}` into `{{.Target}}`\n"

type MergeRequestEvent struct {
	ObjAttr     ObjectAttributes `json:"object_attributes"`
	ProjectInfo Project          `json:"project"`
}

func (h *hook) MergeRequestEvent(b []byte) {
	var mr MergeRequestEvent
	err := json.Unmarshal(b, &mr)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	if mr.ObjAttr.Action == "open" || mr.ObjAttr.Action == "reopen" {
		activeMR(mr, h)
	} else if mr.ObjAttr.Action == "merge" || mr.ObjAttr.Action == "close" {
		deactiveMR(mr, h)
	} else {
		logrus.Infoln("action is NOT one of open, reopen, merge or close")
		logrus.Debugf("action: %v", mr.ObjAttr.Action)
		return
	}
}

func activeMR(mr MergeRequestEvent, h *hook) {
	// if author and assignee are the same person, do nothing
	if mr.ObjAttr.AuthorID == mr.ObjAttr.AssigneeID {
		logrus.Infoln("author and assignee are the same person")
		return
	}

	author, err := h.db.GetUserByID(mr.ObjAttr.AuthorID)
	if err != nil {
		return
	}
	assignee, err := h.db.GetUserByID(mr.ObjAttr.AssigneeID)
	if err != nil {
		return
	}

	// get target Slack channel, priority: description > project > user > #general
	// get from mr description
	var channel string
	re, err := regexp.Compile("/gitlack:\\s?\\S+")
	if err != nil {
		logrus.Errorln(err)
		return
	}
	parsed := re.FindString(mr.ObjAttr.Description)
	if parsed != "" {
		channel = strings.TrimSpace(strings.Replace(parsed, "/gitlack:", "", 1))
	}

	// get from project default channel
	if channel == "" {
		proejct, err := h.db.GetProjectByID(mr.ProjectInfo.ID)
		if err != nil {
			return
		}
		if proejct.DefaultChannel != "" {
			channel = proejct.DefaultChannel
		}

	}

	// get from assignee default channel
	if channel == "" {
		if assignee.DefaultChannel != "" {
			channel = assignee.DefaultChannel
		}
	}

	if channel == "" {
		channel = "general"
	}

	// if user doesn't exist in Slack, use the name of user in GitLab instead
	authorID := author.SlackID
	if author.SlackID == "" {
		authorID = author.Name
	}
	assigneeID := assignee.SlackID
	if assignee.SlackID == "" {
		assigneeID = assignee.Name
	}

	// prepare slack text
	data := map[string]interface{}{
		"Assignee": assigneeID,
		"Author":   authorID,
		"Path":     mr.ProjectInfo.PathWithNamespace,
		"Title":    mr.ObjAttr.Title,
		"Source":   mr.ObjAttr.SourceBranch,
		"Target":   mr.ObjAttr.TargetBranch,
		"Link":     mr.ObjAttr.ObjectURL,
		"MRNum":    mr.ObjAttr.ObjectNum,
	}
	t, err := template.New("slack").Parse(mrTemplate)
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
	smr, err := h.s.PostSlackMessage(channel, slackText.String(), nil)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	// insert new merge request
	newMR := &model.MergeRequest{
		ProjectID:       mr.ProjectInfo.ID,
		MergeRequestNum: mr.ObjAttr.ObjectNum,
		ThreadTS:        smr.TS,
		Channel:         smr.Channel,
	}
	err = h.db.CreateMergeRequest(newMR)
	if err != nil {
		return
	}
}

func deactiveMR(mr MergeRequestEvent, h *hook) {
	mrThread, err := h.db.GetMergeRequest(mr.ProjectInfo.ID, mr.ObjAttr.ObjectNum)
	if err != nil {
		return
	}
	slackText := "This merge request has been "
	if mr.ObjAttr.Action == "merge" {
		slackText += "merged."
	} else {
		slackText += "closed."
	}

	h.s.PostSlackMessage(mrThread.Channel, slackText, nil, mrThread.ThreadTS)
}
