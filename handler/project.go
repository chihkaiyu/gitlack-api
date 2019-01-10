package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (r *router) GetProject(c *gin.Context) {
	namespace := c.Param("namespace")
	path := c.Param("path")
	pathWithNamespace := namespace + path

	p, err := r.db.GetProjectByPath(pathWithNamespace)
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			c.JSON(http.StatusNotFound, gin.H{
				"ok":    false,
				"error": "Project not found",
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
		"ok": true,
		"project": map[string]interface{}{
			"gitlab_id":       p.ID,
			"name":            p.Name,
			"default_channel": p.DefaultChannel,
		},
	})
}

func (r *router) UpdateProject(c *gin.Context) {
	defaultChannel := c.Query("default_channel")
	if defaultChannel == "" {
		logrus.Debugln("Default channel not found")
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Invalid \"default_channel\": %q", defaultChannel),
		})
		return
	}

	// check project exists
	namespace := c.Param("namespace")
	path := c.Param("path")
	pathWithNamespace := namespace + path
	_, err := r.db.GetProjectByPath(pathWithNamespace)
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			c.JSON(http.StatusNotFound, gin.H{
				"ok":    false,
				"error": "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Server error",
		})
		return
	}

	// update project default channel
	err = r.db.UpdateProjectDefaultChannel(pathWithNamespace, defaultChannel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Server error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": fmt.Sprintf("Project: %v updated", pathWithNamespace),
	})
}

func (r *router) WrapSyncProject(c *gin.Context) {
	err := r.SyncProject()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "All projects are synchronized",
	})

}

func (r *router) SyncProject() error {
	projects, err := r.g.GetProject()
	if err != nil {
		return err
	}

	var failed []string
	for _, p := range projects {
		err := r.db.CreateProject(p)
		if err != nil {
			failed = append(failed, p.Name)
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
