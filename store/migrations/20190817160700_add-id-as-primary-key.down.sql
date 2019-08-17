CREATE TABLE TempMergeRequest(
    project_id INTEGER,
    mr_num INTEGER,
    thread_ts CHARACTER(32),
    channel CHARACTER(16),
    PRIMARY KEY (project_id, mr_num),
    FOREIGN KEY (project_id) REFERENCES Project(id)
);

INSERT INTO TempMergeRequest (project_id, mr_num, thread_ts, channel)
    SELECT project_id, mr_num, thread_ts, channel FROM MergeRequest;

DROP TABLE MergeRequest;
ALTER TABLE TempMergeRequest RENAME TO MergeRequest;


CREATE TABLE TempIssue(
    project_id INTEGER,
    issue_num INTEGER,
    thread_ts CHARACTER(32),
    channel CHARACTER(16),
    PRIMARY KEY (project_id, issue_num),
    FOREIGN KEY (project_id) REFERENCES Project(id)
);

INSERT INTO TempIssue (project_id, issue_num, thread_ts, channel)
    SELECT project_id, issue_num, thread_ts, channel FROM Issue;

DROP TABLE Issue;
ALTER TABLE TempIssue RENAME TO Issue;
