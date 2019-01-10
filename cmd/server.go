package main

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"

	"github.com/urfave/cli"
)

func start(c *cli.Context) {
	// debug level if requested by user
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
		gin.SetMode(gin.DebugMode)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
		gin.SetMode(gin.ReleaseMode)
	}

	err := checkFlags(c)
	if err != nil {
		logrus.Fatalln(err)
	}

	// initialize server and required component
	checkMigration(c)
	srv := newServer(c)
	srv.setupRouter()
	srv.setupAndStartCronjob()

	// sync users, projects
	srv.router.SyncUser()
	srv.router.SyncProject()

	addr := c.String("server-addr")
	srv.engine.Run(addr)
}

func checkFlags(c *cli.Context) error {
	required := []string{"slack-token", "gitlab-token"}
	var failed []string
	for _, flag := range required {
		if f := c.String(flag); f == "" {
			failed = append(failed, flag)
		}
	}
	if len(failed) != 0 {
		return errors.New("Missing flags: " + strings.Join(failed, ", "))
	}
	return nil
}
