package webhook

// Project represents the data structure of `project` in GitLab webhook request
type Project struct {
	PathWithNamespace string `json:"path_with_namespace"`
	ID                int    `json:"id"`
	WebURL            string `json:"web_url"`
}

// ObjectAttributes represents the data structure of `object_attributes` in GitLab webhook request
type ObjectAttributes struct {
	Action       string `json:"action"`
	Title        string `json:"title"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	ObjectURL    string `json:"url"`
	Description  string `json:"description"`
	AuthorID     int    `json:"author_id"`
	AssigneeID   int    `json:"assignee_id"`
	ObjectNum    int    `json:"iid"`
	Note         string `json:"note"`
	NoteableType string `json:"noteable_type"`
}

// Issue represents the data structure of issue in comment events
type Issue struct {
	Num int `json:"iid"`
}

// MergeRequest represents the data structure of merge request in comment events
type MergeRequest struct {
	Num int `json:"iid"`
}
