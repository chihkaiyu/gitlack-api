package slack

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"gitlack/model"

	"github.com/sirupsen/logrus"
)

// MessageResponse represents the response of posting message to Slack
type MessageResponse struct {
	OK      bool   `json:"ok"`
	Channel string `json:"channel"`
	TS      string `json:"ts"`
	Err     string `json:"error"`
}

// AttachmentColor is the color on the left side of attachment
const AttachmentColor = "#FF5511"

// Attachment represents the Slack attachment
// see: https://api.slack.com/docs/message-attachments
type Attachment struct {
	Color string `json:"color"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

func (s *slack) PostSlackMessage(channel, text string, author *model.User, atm *Attachment, thread ...string) (*MessageResponse, error) {
	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	reqBody := map[string]string{
		"token":   s.SlackToken,
		"channel": channel,
		"text":    text,
	}
	if len(thread) != 0 {
		reqBody["thread_ts"] = thread[0]
	}
	if atm != nil {
		// attachments shoule be array of structured attachments
		marshaledAtm, err := json.Marshal([]*Attachment{atm})
		if err != nil {
			logrus.Errorln(err)
			return nil, err
		}
		reqBody["attachments"] = string(marshaledAtm)
	}

	if author != nil {
		reqBody["username"] = author.Name + " (Gitlack)"
		reqBody["icon_url"] = author.AvatarURL
	}

	url := s.SlackAPI + "/chat.postMessage"

	res, err := s.client.Post(url, header, nil, reqBody)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err := fmt.Errorf("HTTP response error: %v", string(body))
		logrus.Errorln(err)
		return nil, err
	}

	var smr MessageResponse
	err = json.Unmarshal(body, &smr)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	if !smr.OK {
		errMsg := fmt.Sprintf("Invalid Slack API: %v", smr.Err)
		logrus.Errorln(errMsg)
		return nil, errors.New(errMsg)
	}

	return &smr, nil
}
