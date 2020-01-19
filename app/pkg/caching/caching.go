package caching

import (
	"fmt"
	externalCache "github.com/patrickmn/go-cache"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"sync"
	"time"
)

const cacheKey = "GitlabProjects"

// Errors
const (
	cacheNoObject    = "cache doesn't contains object"
	cacheInvalidType = "cached object type is invalid: %T"
)

type ProjectsCache interface {
	GetProjects() (projects []contracts.Project, exists bool)
	SetProjects(projects []contracts.Project)
	UpdatePipeline(pipelinePush contracts.PipelinePush) error
}

type cache struct {
	*externalCache.Cache
	*sync.Mutex
}

func New(defaultExpiration time.Duration) ProjectsCache {
	result := new(cache)
	result.Cache = externalCache.New(defaultExpiration, 0)
	result.Mutex = new(sync.Mutex)
	return result
}

func (c *cache) GetProjects() (projects []contracts.Project, exists bool) {
	c.Lock()
	defer c.Unlock()
	cached, exists := c.Get(cacheKey)
	if exists {
		projects = cached.([]contracts.Project)
	}
	return
}

func (c *cache) SetProjects(projects []contracts.Project) {
	c.Lock()
	defer c.Unlock()
	c.SetDefault(cacheKey, projects)
}

func (c *cache) UpdatePipeline(pipelinePush contracts.PipelinePush) (err error) {
	c.Lock()
	defer c.Unlock()
	cached, expiration, exists := c.GetWithExpiration(cacheKey)
	if !exists {
		return fmt.Errorf(cacheNoObject)
	}
	projects, ok := cached.([]contracts.Project)
	if !ok {
		return fmt.Errorf(cacheInvalidType, cached)
	}
	for i := 0; i < len(projects); i++ {
		if projects[i].Id == pipelinePush.Project.Id {
			pipelineExists := false
			for j := 0; j < len(projects[i].Pipelines); j++ {
				if projects[i].Pipelines[j].Id == pipelinePush.Attributes.Id {
					projects[i].Pipelines[j].Status = pipelinePush.Attributes.Status
					pipelineExists = true
				}
			}
			if !pipelineExists {
				newPipeline := contracts.Pipeline{
					Id:     pipelinePush.Attributes.Id,
					Sha:    pipelinePush.Attributes.Sha,
					Branch: pipelinePush.Attributes.Branch,
					Status: pipelinePush.Attributes.Status,
					WebUrl: pipelinePush.Commit.Url,
					Commit: &contracts.Commit{
						Title:     pipelinePush.Commit.Message,
						CreatedAt: pipelinePush.Commit.Timestamp,
						Author:    pipelinePush.Commit.Author.Name,
					},
				}
				projects[i].Pipelines = append(projects[i].Pipelines, newPipeline)
			}
			ttl := expiration.Sub(time.Now())
			c.Set(cacheKey, projects, ttl)
			return
		}
	}
	return
}
