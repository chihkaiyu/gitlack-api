package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

// ResponseMetadata represents the link of next cursor
type ResponseMetadata struct {
	NextCursor string `json:"next_cursor"`
}

// Profile is the field represents the personal profile of Slack user
type Profile struct {
	Email     string `json:"email"`
	AvatarURL string `json:"image_72"`
}

// Member is the field represents the information about Slack user
type Member struct {
	ID      string  `json:"id"`
	IsBot   bool    `json:"is_bot"`
	Deleted bool    `json:"deleted"`
	Profile Profile `json:"profile"`
}

// SlackUserResponse is the response of getting Slack user list
type SlackUserResponse struct {
	OK               bool             `json:"ok"`
	Error            string           `json:"error"`
	ResponseMetadata ResponseMetadata `json:"response_metadata"`
	Members          []Member         `json:"members"`
}

type SlackUser struct {
	ID        string
	Email     string
	AvatarURL string
}

func (s *slack) GetUser() ([]*SlackUser, error) {
	url := s.SlackAPI + "/users.list"
	params := map[string]string{
		"limit": "100",
		"token": s.SlackToken,
	}
	var allUsers []*SlackUser
	// run at most 100 times for preventing from infinite loop
	for i := 0; i < 100; i++ {
		res, err := s.client.Get(url, nil, params, nil)
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		if err != nil {
			logrus.Errorln(err)
			return nil, err
		}

		if res.StatusCode != 200 {
			err := fmt.Errorf("HTTP response error: %v", string(body))
			logrus.Errorln(err)
			return nil, err
		}

		var slackResponse SlackUserResponse
		err = json.Unmarshal(body, &slackResponse)
		if err != nil {
			logrus.Errorln(err)
			return nil, err
		}

		if !slackResponse.OK {
			err := fmt.Errorf("Invalid Slack API: %v", slackResponse.Error)
			logrus.Errorln(err)
			return nil, err
		}

		for _, u := range slackResponse.Members {
			if u.IsBot || u.Deleted {
				continue
			}
			slackUser := &SlackUser{
				ID:        u.ID,
				Email:     u.Profile.Email,
				AvatarURL: u.Profile.AvatarURL,
			}
			allUsers = append(allUsers, slackUser)
		}

		// check next page
		if slackResponse.ResponseMetadata.NextCursor == "" {
			break
		}
		params["cursor"] = slackResponse.ResponseMetadata.NextCursor
	}
	return allUsers, nil
}
