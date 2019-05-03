package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gitlack/handler"
)

type server struct {
	engine  *gin.Engine
	router  handler.Handler
	cronjob *cron.Cron
}

func newServer(c *cli.Context) *server {
	return &server{
		engine:  gin.Default(),
		router:  handler.NewHandler(c),
		cronjob: cron.New(),
	}
}

func (s *server) setupRouter() {
	// GitLab webhook
	s.engine.POST("/", s.router.Webhook)

	user := s.engine.Group("/api/user")
	{
		user.GET("/:email", s.router.GetUser)
		user.PUT("/:email", s.router.UpdateUser)
		user.POST("", s.router.WrapSyncUser)
	}

	group := s.engine.Group("/api/group")
	{
		group.PUT("/:namespace/*path", s.router.UpdateGroup)
	}

	project := s.engine.Group("/api/project")
	{
		project.GET("/:namespace/*path", s.router.GetProject)
		project.PUT("/:namespace/*path", s.router.UpdateProject)
		project.POST("", s.router.WrapSyncProject)
	}
}

func (s *server) setupAndStartCronjob() {
	err := s.cronjob.AddFunc("@midnight", func() {
		s.router.SyncUser()
		s.router.SyncProject()
	})
	if err != nil {
		logrus.Errorf("cronjob starting failed: %v", err)
		return
	}
	s.cronjob.Start()
}

func checkMigration(c *cli.Context) error {
	logrus.Debugf("location of migrations folder: %v", c.String("database-migrations"))
	m, err := migrate.New(fmt.Sprintf("file://%v", c.String("database-migrations")), fmt.Sprintf("sqlite3://%v", c.String("database-config")))
	if err != nil {
		logrus.Fatalln(err)
	}
	defer m.Close()

	v, dirty, _ := m.Version()
	logrus.Debugf("Version: %v, Dirty: %v", v, dirty)

	logrus.Infoln("Migrating database...")
	err = m.Up()
	if err != nil && err.Error() != migrate.ErrNoChange.Error() {
		logrus.Fatalln(err)
	}

	return nil
}
