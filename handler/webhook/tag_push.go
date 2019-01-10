package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

const tagPushTemplate = "<@{{.Author}}> has pushed a new tag: <{{.Link}}|{{.Tag}}> to `{{.Path}}`!\n" +
	"{{.Note}}\n"

// TagPushEvent represents the data structure of tag push in GitLab webhook request
type TagPushEvent struct {
	CheckoutSHA string  `json:"checkout_sha"`
	ProjectInfo Project `json:"project"`
	Message     string  `json:"message"`
	AuthorID    int     `json:"user_id"`
}

func (h *hook) TagPushEvent(b []byte) {
	var tagPushInfo TagPushEvent
	err := json.Unmarshal(b, &tagPushInfo)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	// if a tag is deleted, the `checkout_sha` will be Null
	if tagPushInfo.CheckoutSHA == "" {
		logrus.Infoln("tag is deleted")
		return
	}

	// get author
	author, err := h.db.GetUserByID(tagPushInfo.AuthorID)
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
	parsed := re.FindString(tagPushInfo.Message)
	if parsed != "" {
		channel = strings.TrimSpace(strings.Replace(parsed, "/gitlack:", "", 1))
	}

	// get from project default channel
	if channel == "" {
		proejct, err := h.db.GetProjectByID(tagPushInfo.ProjectInfo.ID)
		if err != nil {
			return
		}
		if proejct.DefaultChannel != "" {
			channel = proejct.DefaultChannel
		}
	}

	// get from author default channel
	if channel == "" {
		if author.DefaultChannel != "" {
			channel = author.DefaultChannel
		}
	}

	if channel == "" {
		channel = "general"
	}

	// get tag list
	tagList, err := h.g.GetTagList(tagPushInfo.ProjectInfo.ID)
	tagName := tagList[0].Name
	tagReleaseNote := tagList[0].ReleaseInfo.Description
	tagURL := fmt.Sprintf("%v/tags/%v", tagPushInfo.ProjectInfo.WebURL, tagName)

	// prepare slack text
	data := map[string]interface{}{
		"Author": author.SlackID,
		"Tag":    tagName,
		"Path":   tagPushInfo.ProjectInfo.PathWithNamespace,
		"Note":   tagReleaseNote,
		"Link":   tagURL,
	}
	t, err := template.New("slack").Parse(tagPushTemplate)
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
	_, err = h.s.PostSlackMessage(channel, slackText.String(), nil)
}
