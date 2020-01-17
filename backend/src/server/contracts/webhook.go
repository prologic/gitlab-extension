package contracts

type attributes struct {
	Id         int64    `json:"id"`
	Branch     string   `json:"ref"`
	Sha        string   `json:"sha"`
	Status     string   `json:"status"`
	Stages     []string `json:"stages"`
	CreatedAt  string   `json:"created_at"`
	FinishedAt string   `json:"finished_at"`
	Duration   int64    `json:"duration"`
}

type user struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

type project struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Namespace   string `json:"namespace"`
	WebUrl      string `json:"web_url"`
}

type commit struct {
	Id        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Url       string `json:"url"`
	Author    author `json:"author"`
}

type author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type build struct {
	Id         int64     `json:"id"`
	Stage      string    `json:"stage"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	CreatedAt  string    `json:"created_at"`
	StartedAt  string    `json:"started_at"`
	FinishedAt string    `json:"finished_at"`
	When       string    `json:"when"`
	Manual     bool      `json:"manual"`
	User       user      `json:"user"`
	Runner     runner    `json:"runner"`
	Artifacts  artifacts `json:"artifacts_file"`
}

type runner struct {
	Id          int64  `json:"id"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	IsShared    bool   `json:"is_shared"`
}

type artifacts struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

type PipelinePush struct {
	Kind       string     `json:"object_kind"`
	Attributes attributes `json:"object_attributes"`
	User       user       `json:"user"`
	Project    project    `json:"project"`
	Commit     commit     `json:"commit"`
	Builds     []build    `json:"builds"`
}
