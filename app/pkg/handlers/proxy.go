package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/caching"
	"github.com/ricdeau/gitlab-extension/app/pkg/config"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"github.com/ricdeau/gitlab-extension/app/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	privateToken    = "Private-Token"
	pipelinesNumber = 5
)

// ProxyHandler that performs multiple requests to gitlab API and returns single combined response.
// with all projects, first N pipelines for each project, and last commit for each pipeline.
type ProxyHandler struct {
	*config.Config
	Log            *logrus.Logger
	gitlabApiV4Url string
	client         *http.Client
	cache          caching.ProjectsCache
}

// Create new instance of ProxyHandler.
// config - Global config
// cache - Caching module
// logger - Logging module
func NewProxyHandler(conf *config.Config, cache caching.ProjectsCache, logger *logrus.Logger) *ProxyHandler {
	instance := &ProxyHandler{}
	instance.Config = conf
	instance.cache = cache
	instance.Log = logger
	instance.gitlabApiV4Url = fmt.Sprintf("%s/api/v4/", conf.GitlabUri)
	instance.client = &http.Client{
		Timeout: time.Second * 30,
	}
	return instance
}

// Handle '/projects' request
func (handler *ProxyHandler) Handle(c *gin.Context) {

	corrId, exists := c.Get(logging.CorrelationIdKey)
	if !exists {
		corrId = uuid.New()
	}
	logger := handler.Log.WithField(logging.CorrelationIdKey, corrId)

	// parse project ids
	projectIdsParam := c.Query("project_ids")
	var projectIds []int64
	if projectIdsParam != "" {
		for _, idParam := range strings.Split(projectIdsParam, " ") {
			id, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				logger.Warnf("Invalid projectId %s will be skipped", idParam)
				continue
			}
			projectIds = append(projectIds, id)
		}
	}

	// parse branches
	branchesParam := c.Query("branches")
	var branches []string
	if branchesParam != "" {
		for _, branchParam := range strings.Split(branchesParam, " ") {
			branches = append(branches, branchParam)
		}
	}

	projects, err := handler.getProjects(pipelinesNumber, logger)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err,
		})
		return
	}

	// filter projects by ids and pipelines by branches
	projects = handler.filterProjects(projects, projectIds, branches)

	c.JSON(200, gin.H{
		"projects": projects,
	})
}

// Filter projects by provided ids and filter each project pipelines by provided branches
// projects - Projects structure from gitlab API
// projectIds - Ids of gitlab projects to filter
// branches - Branches to filter pipelines
func (handler *ProxyHandler) filterProjects(
	projects []contracts.Project, projectIds []int64, branches []string) (result []contracts.Project) {
	//
	//for _, proj := range projects {
	//	if funk.Any(projectIds) && !funk.Contains(projectIds, proj.Id) {
	//		continue
	//	}
	//	var filteredPipelines []contracts.Pipeline
	//	for _, pipe := range proj.Pipelines {
	//		if funk.Any(branches) && !funk.Contains(branches, pipe.Branch) {
	//			continue
	//		}
	//		filteredPipelines = append(filteredPipelines, pipe)
	//	}
	//	proj.Pipelines = filteredPipelines
	//	result = append(result, proj)
	//}
	return
}

// Gets all projects, allowed for private token that provided through ProxyHandler.Config.
// Projects will be cached, if cache is empty or expired, http request will be processed.
// nPipelines - top N pipelines to take
func (handler *ProxyHandler) getProjects(
	nPipelines int,
	logger *logrus.Entry) (result []contracts.Project, err error) {

	// return if cached
	exists := false
	result, exists = handler.cache.GetProjects()
	if exists {
		return
	}

	// get new if not found in cache
	url := fmt.Sprintf("%s/projects", handler.gitlabApiV4Url)
	response, err := handler.performGetRequest(url, logger)
	if err != nil {
		return
	}
	defer response.Body.Close()
	var rawJson []map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&rawJson)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	ch := make(chan contracts.Project, 4)
	for _, projRaw := range rawJson {
		wg.Add(1)
		go func(p map[string]interface{}) {
			defer wg.Done()
			project := contracts.Project{
				Id:           int64(p["id"].(float64)),
				Name:         p["name"].(string),
				Namespace:    p["namespace"].(map[string]interface{})["name"].(string),
				LastActivity: p["last_activity_at"].(string),
				WebUrl:       p["web_url"].(string),
			}
			// add pipelines to project
			project.Pipelines, err = handler.getPipelines(project.Id, nPipelines, logger)
			if err != nil {
				logger.Errorf("Error while getting pipelines: %v", err)
			}
			ch <- project
		}(projRaw)
	}

	for range rawJson {
		result = append(result, <-ch)
	}
	wg.Wait()

	handler.cache.SetProjects(result)

	return
}

// Gets pipelines for project.
// projectId - the identifier of gitlab project
// nPipelines - top N pipelines to take
func (handler *ProxyHandler) getPipelines(
	projectId int64,
	nPipelines int,
	logger *logrus.Entry) (pipelines []contracts.Pipeline, err error) {

	url := fmt.Sprintf("%s/projects/%d/pipelines", handler.gitlabApiV4Url, projectId)
	response, err := handler.performGetRequest(url, logger)
	if err != nil {
		return
	}
	defer response.Body.Close()
	var rawJson []map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&rawJson)
	if err != nil {
		return
	}
	if nPipelines > len(rawJson) {
		nPipelines = len(rawJson)
	}
	for _, p := range rawJson[:nPipelines] {
		pipeline := contracts.Pipeline{
			Id:     int64(p["id"].(float64)),
			Sha:    p["sha"].(string),
			Branch: p["ref"].(string),
			Status: p["status"].(string),
			WebUrl: p["web_url"].(string),
		}
		// add last commit to pipeline
		pipeline.Commit, err = handler.getCommitForProject(projectId, pipeline.Sha, logger)
		if err != nil {
			return
		}
		pipelines = append(pipelines, pipeline)
	}
	return
}

// Gets commit and converts it to contracts.Commit struct.
// projectId - the identifier of gitlab project
// sha - commit's SHA
func (handler *ProxyHandler) getCommitForProject(
	projectId int64,
	sha string,
	logger *logrus.Entry) (result contracts.Commit, err error) {

	url := fmt.Sprintf("%s/projects/%d/repository/commits/%s", handler.gitlabApiV4Url, projectId, sha)
	response, err := handler.performGetRequest(url, logger)
	if err != nil {
		return
	}
	defer response.Body.Close()
	var rawJson map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&rawJson)
	if err != nil {
		return
	}
	result = contracts.Commit{
		Title:     rawJson["title"].(string),
		CreatedAt: rawJson["created_at"].(string),
		Author:    rawJson["author_name"].(string),
	}
	return
}

// Performs GET request with Private-Token header and returns response.
// url - request's url
func (handler *ProxyHandler) performGetRequest(url string, logger *logrus.Entry) (*http.Response, error) {
	headers := map[string]string{privateToken: handler.GitlabToken}
	return utils.PerformGetRequest(handler.client, url, headers, logger)
}
