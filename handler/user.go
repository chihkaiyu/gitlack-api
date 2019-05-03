package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gitlack/model"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (r *router) GetUser(c *gin.Context) {
	email := c.Param("email")

	u, err := r.db.GetUserByEmail(email)
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			c.JSON(http.StatusNotFound, gin.H{
				"ok":    false,
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Server error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"user": u,
	})
}

func (r *router) UpdateUser(c *gin.Context) {
	defaultChannel := c.Query("default_channel")
	if defaultChannel == "" {
		logrus.Debugln("Default channel not found")
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Invalid \"default_channel\": %q", defaultChannel),
		})
		return
	}

	// check user exists
	email := c.Param("email")
	_, err := r.db.GetUserByEmail(email)
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			c.JSON(http.StatusNotFound, gin.H{
				"ok":    false,
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Server error",
		})
		return
	}

	// update user default channel
	err = r.db.UpdateUserDefaultChannel(email, defaultChannel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Server error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": fmt.Sprintf("User: %v updated", email),
	})
}

func (r *router) WrapSyncUser(c *gin.Context) {
	err := r.SyncUser()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "All users are synchronized",
	})
}

func (r *router) SyncUser() error {
	gitlabUsers, err := r.g.GetUser()
	if err != nil {
		return err
	}

	slackUsers, err := r.s.GetUser()
	if err != nil {
		return err
	}

	// join GitLab and Slack user by email (drop domain part)
	// drop the users not exist in GitLab
	combinedUsers := make(map[string]*model.User)
	for _, g := range gitlabUsers {
		email := strings.Split(g.Email, "@")[0]
		u := &model.User{
			Email:     email,
			SlackID:   "",
			GitLabID:  g.ID,
			Name:      g.Name,
			AvatarURL: "",
		}
		combinedUsers[g.Email] = u
	}

	for _, s := range slackUsers {
		if u, exist := combinedUsers[s.Email]; exist {
			u.SlackID = s.ID
			u.AvatarURL = s.AvatarURL
			combinedUsers[s.Email] = u
		}
	}

	var failed []string
	for _, u := range combinedUsers {
		err := r.db.CreateUser(u)
		if err != nil {
			failed = append(failed, u.Email)
		}
	}
	if len(failed) != 0 {
		errMsg, err := json.Marshal(failed)
		if err != nil {
			logrus.Errorln(err)
			return err
		}
		return errors.New(string(errMsg))
	}
	return nil
}
