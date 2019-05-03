/*
Sqlite has no way to remove column directly.
  1. create new table.
  2. copy all data,
  3. drop old table,
  4. rename the new one.
ref:
  * https://stackoverflow.com/questions/8442147/how-to-delete-or-add-column-in-sqlite 
  * https://www.sqlite.org/lang_altertable.html
*/
CREATE TABLE "TempUserTable" (
	"gitlab_id"	INT,
	"email"	VARCHAR(255) NOT NULL,
	"slack_id"	VARCHAR(9) NOT NULL,
	"name"	VARCHAR(255) NOT NULL,
	"default_channel"	VARCHAR(32) DEFAULT '',
	PRIMARY KEY("gitlab_id")
);

INSERT INTO "main"."TempUserTable"
("default_channel","email","gitlab_id","name","slack_id")
SELECT "default_channel","email","gitlab_id","name","slack_id" FROM "main"."User";

DROP TABLE "main"."User";
ALTER TABLE "main"."TempUserTable" RENAME TO "User"