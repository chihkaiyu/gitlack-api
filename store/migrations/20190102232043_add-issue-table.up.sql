CREATE TABLE IF NOT EXISTS Issue(
    project_id INTEGER,
    issue_num INTEGER,
    thread_ts CHARACTER(32),
    channel CHARACTER(16),
    PRIMARY KEY (project_id, issue_num),
    FOREIGN KEY (project_id) REFERENCES Project(id)
);