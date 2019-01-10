CREATE TABLE IF NOT EXISTS MergeRequest(
    project_id INTEGER,
    mr_num INTEGER,
    thread_ts CHARACTER(32),
    channel CHARACTER(16),
    PRIMARY KEY (project_id, mr_num),
    FOREIGN KEY (project_id) REFERENCES Project(id)
);