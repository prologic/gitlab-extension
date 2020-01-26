package contracts

type ProjectsResponse struct {
	Projects []Project `json:"projects"`
}

type Project struct {
	Id           int64      `json:"id"`
	Name         string     `json:"name"`
	Namespace    string     `json:"namespace"`
	LastActivity string     `json:"last_activity"`
	WebUrl       string     `json:"web_url"`
	Pipelines    []Pipeline `json:"pipelines"`
}

type Pipeline struct {
	Id     int64   `json:"id"`
	Sha    string  `json:"sha"`
	Branch string  `json:"branch"`
	Status string  `json:"status"`
	WebUrl string  `json:"web_url"`
	Commit *Commit `json:"PipelineCommit"`
}

type Commit struct {
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	Author    string `json:"Author"`
}

func NewProjectsResponse(projects []Project) ProjectsResponse {
	return ProjectsResponse{projects}
}
