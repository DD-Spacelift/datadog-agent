// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build docker

package ecs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	workloadmeta "github.com/DataDog/datadog-agent/comp/core/workloadmeta/def"
	v1 "github.com/DataDog/datadog-agent/pkg/util/ecs/metadata/v1"
	"github.com/DataDog/datadog-agent/pkg/util/ecs/metadata/v3or4"
)

// TestPullWithV1Parser tests the collector's Pull method by setting the taskCollectionParser to parseTasksFromV1Endpoint
// which is the default parser when other metadata endpoints are not available.
func TestPullWithV1Parser(t *testing.T) {
	entityID := "task1"
	containerID := "someID"
	tags := map[string]string{"foo": "bar"}

	tests := []struct {
		name                string
		collectResourceTags bool
		expectedTags        map[string]string
	}{
		{
			name:                "collect tags",
			collectResourceTags: true,
			expectedTags:        tags,
		},
		{
			name:                "don't collect tags",
			collectResourceTags: false,
			expectedTags:        nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := collector{
				resourceTags: make(map[string]resourceTags),
				seen:         make(map[workloadmeta.EntityID]struct{}),
			}

			c.metaV1 = &fakev1EcsClient{
				mockGetTasks: func(_ context.Context) ([]v1.Task, error) {
					return []v1.Task{
						{
							Arn: entityID,
							Containers: []v1.Container{
								{DockerID: containerID},
							},
						},
					}, nil
				},
			}
			c.store = &fakeWorkloadmetaStore{}
			c.metaV3or4 = func(_, _ string) v3or4.Client {
				return &fakev3or4EcsClient{
					mockGetTaskWithTags: func(context.Context) (*v3or4.Task, error) {
						return &v3or4.Task{
							TaskTags: map[string]string{
								"foo": "bar",
							},
						}, nil
					},
				}
			}

			c.hasResourceTags = true
			c.collectResourceTags = test.collectResourceTags
			c.taskCollectionParser = c.parseTasksFromV1Endpoint

			err := c.Pull(context.TODO())
			require.NoError(t, err)

			taskTags := c.resourceTags[entityID].tags
			assert.Equal(t, taskTags, test.expectedTags)

			// This is only needed because of the workaround about the empty
			// runtime documented in the parseTaskContainers function. Remove
			// this when the workaround is no longer needed.
			storedContainer, err := c.store.GetContainer(containerID)
			require.NoError(t, err)
			assert.Empty(t, storedContainer.Runtime)
		})
	}

}
