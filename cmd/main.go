package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

var rootDir, _ = os.Getwd()

var flags = []cli.Flag{
	cli.BoolFlag{
		Name:  "debug",
		Usage: "enable server debug mode",
	},
	cli.StringFlag{
		EnvVar: "SLACK_SCHEME",
		Name:   "slack-scheme",
		Usage:  "Slack scheme, http or https, default https",
		Value:  "https",
	},
	cli.StringFlag{
		EnvVar: "SLACK_DOMAIN",
		Name:   "slack-domain",
		Usage:  "Slack domain, default slack.com",
		Value:  "slack.com",
	},
	cli.StringFlag{
		EnvVar: "SLACK_TOKEN",
		Name:   "slack-token",
		Usage:  "token for accessing Slack",
	},
	cli.StringFlag{
		EnvVar: "GITLAB_SCHEME",
		Name:   "gitlab-scheme",
		Usage:  "GitLab scheme, http or https, default https",
		Value:  "https",
	},
	cli.StringFlag{
		EnvVar: "GITLAB_DOMAIN",
		Name:   "gitlab-domain",
		Usage:  "GitLab domain",
		Value:  "gitlab.com",
	},
	cli.StringFlag{
		EnvVar: "GITLAB_TOKEN",
		Name:   "gitlab-token",
		Usage:  "token for accessing GitLab",
	},
	cli.StringFlag{
		EnvVar: "SERVER_ADDR",
		Name:   "server-addr",
		Usage:  "server address",
		Value:  ":5000",
	},
	cli.StringFlag{
		EnvVar: "DATABASE_CONFIG",
		Name:   "database-config",
		Usage:  "database file path",
		Value:  filepath.Join(rootDir, "db", "gitlack.db"),
	},
	cli.StringFlag{
		EnvVar: "DATABASE_MIGRATIONS",
		Name:   "database-migrations",
		Usage:  "database migrations folder",
		Value:  filepath.Join(rootDir, "store", "migrations"),
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "gitlack"
	app.Version = "v0.0.2"
	app.Usage = "gitlack"
	app.Action = start
	app.Flags = flags

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
