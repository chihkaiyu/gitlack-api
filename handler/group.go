package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (r *router) UpdateGroup(c *gin.Context) {
	defaultChannel := c.Query("default_channel")
	if defaultChannel == "" {
		logrus.Debugln("Default channel not found")
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Invalid \"default_channel\": %q", defaultChannel),
		})
		return
	}

	namespace := c.Param("namespace")
	path := c.Param("path")
	pathWithNamespace := namespace + path

	err := r.db.UpdateGroupDefaultChannel(pathWithNamespace, defaultChannel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": fmt.Sprintf("Group: %v updated", pathWithNamespace),
	})
}
