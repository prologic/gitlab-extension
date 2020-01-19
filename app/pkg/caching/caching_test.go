package caching

import (
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	projectId  int64 = 1
	pipelineId int64 = 31
)

func TestNew(t *testing.T) {
	c := New(-1)
	assert.NotNil(t, c)
	assert.IsType(t, &cache{}, c)
}

func TestCache_SetProjects(t *testing.T) {
	c := New(-1)
	expected := make([]contracts.Project, 0)
	c.SetProjects(expected)

	actual, exists := c.(*cache).Get(cacheKey)
	assert.True(t, exists)
	assert.Equal(t, expected, actual)
}

func TestCache_GetProjects(t *testing.T) {
	c := New(200 * time.Millisecond)
	actual, exists := c.GetProjects()
	assert.False(t, exists)
	assert.Nil(t, actual)

	expected := make([]contracts.Project, 0)
	c.SetProjects(expected)

	actual, exists = c.GetProjects()
	assert.True(t, exists)
	assert.Equal(t, expected, actual)

	time.Sleep(200 * time.Millisecond)

	actual, exists = c.GetProjects()
	assert.False(t, exists)
	assert.Nil(t, actual)
}

func TestCache_UpdatePipeline_ExistingPipeline(t *testing.T) {
	const success = "success"
	c := New(-1)
	before := createProjects(true)
	assert.Len(t, before[0].Pipelines, 1)
	assert.NotEqual(t, success, before[0].Pipelines[0].Status)

	c.SetProjects(before)
	err := c.UpdatePipeline(createTestPipelinePush())
	if assert.NoError(t, err) {
		after, exists := c.GetProjects()
		assert.True(t, exists)
		assert.Len(t, after, 1)
		pipelines := after[0].Pipelines
		assert.Len(t, pipelines, 1)
		assert.Equal(t, success, pipelines[0].Status)
	}
}

func TestCache_UpdatePipeline_NewPipeline(t *testing.T) {
	c := New(-1)
	before := createProjects(false)
	assert.Nil(t, before[0].Pipelines)

	c.SetProjects(before)
	err := c.UpdatePipeline(createTestPipelinePush())
	if assert.NoError(t, err) {
		after, exists := c.GetProjects()
		assert.True(t, exists)
		assert.Len(t, after, 1)
		pipelines := after[0].Pipelines
		assert.Len(t, pipelines, 1)
		assert.Equal(t, pipelineId, pipelines[0].Id)
	}
}

func TestCache_UpdatePipeline_NoObject(t *testing.T) {
	c := New(-1)
	err := c.UpdatePipeline(createTestPipelinePush())
	assert.EqualError(t, err, cacheNoObject)
}

func TestCache_UpdatePipeline_InvalidObjectType(t *testing.T) {
	obj := struct{}{}
	c := New(-1)
	c.(*cache).SetDefault(cacheKey, obj)
	err := c.UpdatePipeline(createTestPipelinePush())
	expectedError := fmt.Sprintf(cacheInvalidType, obj)
	assert.EqualError(t, err, expectedError)
}

func createProjects(withPipeline bool) []contracts.Project {
	var pipelines []contracts.Pipeline
	if withPipeline {
		pipelines = []contracts.Pipeline{
			{
				Id:     pipelineId,
				Sha:    "started",
				Branch: "",
				Status: "",
				WebUrl: "",
				Commit: nil,
			},
		}
	}
	return []contracts.Project{
		{
			Id:           projectId,
			Name:         "Gitlab Test",
			Namespace:    "Gitlab Org",
			LastActivity: "2016-08-12 15:23:28 UTC",
			WebUrl:       "http://192.168.64.1:3005/gitlab-org/gitlab-test",
			Pipelines:    pipelines,
		},
	}
}

func createTestPipelinePush() contracts.PipelinePush {
	user := contracts.User{
		Name:     "Administrator",
		Username: "root",
	}
	return contracts.PipelinePush{
		Kind: "pipeline",
		Attributes: &contracts.Attributes{
			Id:         pipelineId,
			Branch:     "master",
			Sha:        "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
			Status:     "success",
			Stages:     nil,
			CreatedAt:  "2016-08-12 15:23:28 UTC",
			FinishedAt: "2016-08-12 15:26:29 UTC",
			Duration:   63,
		},
		User: &user,
		Project: &contracts.PipelineProject{
			Id:          projectId,
			Name:        "Gitlab Test",
			Description: "Description",
			Namespace:   "Gitlab Org",
			WebUrl:      "http://192.168.64.1:3005/gitlab-org/gitlab-test",
		},
		Commit: &contracts.PipelineCommit{
			Id:        "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
			Message:   "test\n",
			Timestamp: "2016-08-12T17:23:21+02:00",
			Url:       "http://example.com/gitlab-org/gitlab-test/commit/bcbb5ec396a2c0f828686f14fac9b80b780504f2",
			Author: &contracts.Author{
				Name:  "User",
				Email: "user@gitlab.com",
			},
		},
		Builds: []contracts.Build{
			{
				Id:         300,
				Stage:      "deploy",
				Name:       "production",
				Status:     "success",
				CreatedAt:  "2016-08-12 15:23:28 UTC",
				StartedAt:  "2016-08-12 15:26:12 UTC",
				FinishedAt: "",
				When:       "on_success",
				Manual:     false,
				User:       &user,
				Runner: &contracts.Runner{
					Id:          380987,
					Description: "main",
					Active:      true,
					IsShared:    false,
				},
				Artifacts: nil,
			},
		},
	}
}
