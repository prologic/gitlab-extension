package contracts

type Attributes struct {
	Id         int64    `json:"id"`
	Branch     string   `json:"ref"`
	Sha        string   `json:"sha"`
	Status     string   `json:"status"`
	Stages     []string `json:"stages"`
	CreatedAt  string   `json:"created_at"`
	FinishedAt string   `json:"finished_at"`
	Duration   int64    `json:"duration"`
}

type User struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

type PipelineProject struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Namespace   string `json:"namespace"`
	WebUrl      string `json:"web_url"`
}

type PipelineCommit struct {
	Id        string  `json:"id"`
	Message   string  `json:"message"`
	Timestamp string  `json:"timestamp"`
	Url       string  `json:"url"`
	Author    *Author `json:"Author"`
}

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Build struct {
	Id         int64      `json:"id"`
	Stage      string     `json:"stage"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	CreatedAt  string     `json:"created_at"`
	StartedAt  string     `json:"started_at"`
	FinishedAt string     `json:"finished_at"`
	When       string     `json:"when"`
	Manual     bool       `json:"manual"`
	User       *User      `json:"User"`
	Runner     *Runner    `json:"Runner"`
	Artifacts  *Artifacts `json:"artifacts_file"`
}

type Runner struct {
	Id          int64  `json:"id"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	IsShared    bool   `json:"is_shared"`
}

type Artifacts struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

type PipelinePush struct {
	Kind       string           `json:"object_kind"`
	Attributes *Attributes      `json:"object_attributes"`
	User       *User            `json:"User"`
	Project    *PipelineProject `json:"PipelineProject"`
	Commit     *PipelineCommit  `json:"PipelineCommit"`
	Builds     []Build          `json:"builds"`
}
