package caching

import (
	"fmt"
	externalCache "github.com/patrickmn/go-cache"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	cacheKey         = "GitlabProjects"
	UpdateCacheTopic = "cache"
)

type ProjectsCache interface {
	GetProjects() (projects []contracts.Project, exists bool)
	SetProjects(projects []contracts.Project)
	UpdatePipeline(pipelinePush contracts.PipelinePush) (err error)
}

type Cache struct {
	*externalCache.Cache
	*sync.Mutex
	queue *broker.MessageBroker
}

func NewCache(externalCache *externalCache.Cache, queue *broker.MessageBroker, logger *logrus.Logger) *Cache {
	result := Cache{}
	result.Cache = externalCache
	result.queue = queue
	result.Mutex = new(sync.Mutex)
	result.queue.AddTopic(UpdateCacheTopic)
	result.queue.Subscribe(UpdateCacheTopic, func(message interface{}) {
		err := result.UpdatePipeline(message.(contracts.PipelinePush))
		if err != nil {
			logger.Errorf("Error while updating cache: %v", err)
		}
	})
	return &result
}

func (c *Cache) GetProjects() (projects []contracts.Project, exists bool) {
	c.Lock()
	defer c.Unlock()
	cached, exists := c.Get(cacheKey)
	if exists {
		projects = cached.([]contracts.Project)
	}
	return
}

func (c *Cache) SetProjects(projects []contracts.Project) {
	c.Lock()
	defer c.Unlock()
	c.SetDefault(cacheKey, projects)
}

func (c *Cache) UpdatePipeline(pipelinePush contracts.PipelinePush) (err error) {
	c.Lock()
	defer c.Unlock()
	cached, expiration, exists := c.GetWithExpiration(cacheKey)
	if !exists {
		return fmt.Errorf("cache doesn't contains object")
	}
	projects := cached.([]contracts.Project)
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
					Commit: contracts.Commit{
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
