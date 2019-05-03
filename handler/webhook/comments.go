package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/sirupsen/logrus"
)

const commentTemplate = "{{.Author}} has <{{.Link}}|commented:>\n" +
	"{{.Desc}}"

type CommentsEvent struct {
	ObjAttr          ObjectAttributes `json:"object_attributes"`
	ProjectInfo      Project          `json:"project"`
	IssueInfo        Issue            `json:"issue"`
	MergeRequestInfo MergeRequest     `json:"merge_request"`
}

func (h *hook) CommentsEvent(b []byte) {
	var comment CommentsEvent
	err := json.Unmarshal(b, &comment)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	if comment.ObjAttr.NoteableType == "Issue" {
		issuesComment(comment, h)
	} else if comment.ObjAttr.NoteableType == "MergeRequest" {
		mrComment(comment, h)
	} else {
		logrus.Infoln(fmt.Sprintf("comment type not supported: %v", comment.ObjAttr.NoteableType))
		return
	}
}

func issuesComment(comment CommentsEvent, h *hook) {
	// get author of comment
	author, err := h.db.GetUserByID(comment.ObjAttr.AuthorID)
	if err != nil {
		return
	}

	// get issue thread ts
	issue, err := h.db.GetIssue(comment.ProjectInfo.ID, comment.IssueInfo.Num)
	if err != nil {
		return
	}

	// prepare Slack text
	data := map[string]interface{}{
		"Author": author.Name,
		"Link":   comment.ObjAttr.ObjectURL,
		"Desc":   comment.ObjAttr.Note,
	}
	t, err := template.New("slack").Parse(commentTemplate)
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
	h.s.PostSlackMessage(issue.Channel, slackText.String(), author, nil, issue.ThreadTS)
}

func mrComment(comment CommentsEvent, h *hook) {
	// get author of comment
	author, err := h.db.GetUserByID(comment.ObjAttr.AuthorID)
	if err != nil {
		return
	}

	// get issue thread ts
	mr, err := h.db.GetMergeRequest(comment.ProjectInfo.ID, comment.MergeRequestInfo.Num)
	if err != nil {
		return
	}

	// prepare Slack text
	data := map[string]interface{}{
		"Author": author.Name,
		"Link":   comment.ObjAttr.ObjectURL,
		"Desc":   comment.ObjAttr.Note,
	}
	t, err := template.New("slack").Parse(commentTemplate)
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
	h.s.PostSlackMessage(mr.Channel, slackText.String(), nil, nil, mr.ThreadTS)
}
