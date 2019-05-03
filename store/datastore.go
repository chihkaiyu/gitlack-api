package store

import (
	"time"

	"github.com/urfave/cli"

	"github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"gitlack/model"
)

type datastore struct {
	*sqlx.DB
}

// NewStore returns a db connection
func NewStore(c *cli.Context) Store {
	return &datastore{
		DB: open(c.String("database-config")),
	}
}

func open(config string) *sqlx.DB {
	logrus.Debugf("database file location: %v", config)
	db, err := sqlx.Open("sqlite3", config)
	if err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("database connection failed")
	}

	if err := pingDatabase(db); err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("database ping attempts failed")
	}

	return db

}

func pingDatabase(db *sqlx.DB) error {
	var err error
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			return nil
		}
		logrus.Infof("database ping failed. retry in 1s")
		time.Sleep(time.Second)
	}
	return err
}

func (ds *datastore) GetProjectByPath(path string) (*model.Project, error) {
	var p model.Project
	err := ds.Get(&p, "SELECT * FROM Project WHERE name = ?", path)
	if err != nil {
		logrus.Debugf("GetProjectByPath fail, path: %v", path)
		logrus.Errorln(err)
		return nil, err
	}
	return &p, nil
}

func (ds *datastore) GetProjectByID(id int) (*model.Project, error) {
	var p model.Project
	err := ds.Get(&p, "SELECT * FROM Project WHERE id = ?", id)
	if err != nil {
		logrus.Debugf("GetProjectByID fail, id: %v", id)
		logrus.Errorln(err)
		return nil, err
	}
	return &p, nil
}

func (ds *datastore) GetUserByEmail(email string) (*model.User, error) {
	var u model.User
	err := ds.Get(&u, "SELECT * FROM User WHERE email = ?", email)
	if err != nil {
		logrus.Debugf("GetUserByEmail fail, email: %v", email)
		logrus.Errorln(err)
		return nil, err
	}
	return &u, nil
}

func (ds *datastore) GetUserByID(id int) (*model.User, error) {
	var u model.User
	err := ds.Get(&u, "SELECT * FROM User WHERE gitlab_id = ?", id)
	if err != nil {
		logrus.Debugf("GetUserByID fail, id: %v", id)
		logrus.Errorln(err)
		return nil, err
	}
	return &u, nil
}

func (ds *datastore) GetMergeRequest(projectID, mrNum int) (*model.MergeRequest, error) {
	var mr model.MergeRequest
	err := ds.Get(&mr, "SELECT * FROM MergeRequest WHERE project_id = ? and mr_num = ?", projectID, mrNum)
	if err != nil {
		logrus.Debugf("GetMergeRequest fail, projectID: %v, mrNum: %v", projectID, mrNum)
		logrus.Errorln(err)
		return nil, err
	}
	return &mr, nil
}

func (ds *datastore) GetIssue(projectID, issueNum int) (*model.Issue, error) {
	var issue model.Issue
	err := ds.Get(&issue, "SELECT * FROM Issue WHERE project_id = ? and issue_num = ?", projectID, issueNum)
	if err != nil {
		logrus.Debugf("GetIssue fail, projectID: %v, issueNum: %v", projectID, issueNum)
		logrus.Errorln(err)
		return nil, err
	}
	return &issue, nil
}

func (ds *datastore) UpdateUserDefaultChannel(email, channel string) error {
	_, err := ds.Exec("UPDATE User SET default_channel=? WHERE email=?", channel, email)
	if err != nil {
		logrus.Debugf("UpdateUserDefaultChannel fail, email: %v, channel: %v", email, channel)
		logrus.Errorln(err)
		return err
	}
	return nil
}

func (ds *datastore) UpdateProjectDefaultChannel(name, channel string) error {
	_, err := ds.Exec("UPDATE Project SET default_channel=? WHERE name=?", channel, name)
	if err != nil {
		logrus.Debugf("UpdateProjectDefaultChannel fail, name: %v, channel: %v", name, channel)
		logrus.Errorln(err)
		return err
	}
	return nil
}

func (ds *datastore) UpdateGroupDefaultChannel(name, channel string) error {
	_, err := ds.Exec("UPDATE Project SET default_channel=? WHERE name like ?", channel, name+"%")
	if err != nil {
		logrus.Debugf("UpdateGroupDefaultChannel fail, name: %v, channel: %v", name, channel)
		logrus.Errorln(err)
		return err
	}
	return nil
}

func (ds *datastore) CreateUser(u *model.User) error {
	sql := `
INSERT INTO User (gitlab_id, email, slack_id, name, avatar_url)
VALUES (:gitlab_id, :email, :slack_id, :name, :avatar_url)
ON CONFLICT(gitlab_id) DO UPDATE SET email=:email, slack_id=:slack_id, name=:name, avatar_url=:avatar_url
`
	_, err := ds.NamedExec(sql, u)
	if err != nil {
		logrus.Debugf("CreateUser fail, model.User: %v", u)
		logrus.Errorln(err)
		return err
	}
	return nil
}

func (ds *datastore) CreateProject(p *model.Project) error {
	sql := `
INSERT INTO Project (id, name)
VALUES (:id, :name)
ON CONFLICT(id) DO UPDATE SET name=:name
`
	_, err := ds.NamedExec(sql, p)
	if err != nil {
		logrus.Debugf("CreateProject fail, model.Project: %v", p)
		logrus.Errorln(err)
		return err
	}
	return nil
}

func (ds *datastore) CreateMergeRequest(mr *model.MergeRequest) error {
	sql := `
INSERT INTO MergeRequest (project_id, mr_num, thread_ts, channel)
VALUES (:project_id, :mr_num, :thread_ts, :channel)
ON CONFLICT(project_id, mr_num) DO UPDATE SET thread_ts=:thread_ts, channel=:channel
`
	_, err := ds.NamedExec(sql, mr)
	if err != nil {
		logrus.Debugf("CreateMergeRequest fail, model.MergeRequest: %v", mr)
		logrus.Errorln(err)
		return err
	}
	return nil
}

func (ds *datastore) CreateIssue(issue *model.Issue) error {
	sql := `
INSERT INTO Issue (project_id, issue_num, thread_ts, channel)
VALUES (:project_id, :issue_num, :thread_ts, :channel)
ON CONFLICT(project_id, issue_num) DO UPDATE SET thread_ts=:thread_ts, channel=:channel
`
	_, err := ds.NamedExec(sql, issue)
	if err != nil {
		logrus.Debugf("CreateIssue fail, model.Issue: %v", issue)
		logrus.Errorln(err)
		return err
	}
	return nil
}
