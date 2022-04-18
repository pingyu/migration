// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduler

import (
	"github.com/tikv/migration/cdc/cdc/model"
	"github.com/tikv/migration/cdc/cdc/scheduler/util"
)

// InfoProvider is the interface to get information about the internal states of the scheduler.
// We need this interface so that we can provide the information through HTTP API.
type InfoProvider interface {
	// GetTaskStatuses returns the task statuses.
	GetTaskStatuses() (map[model.CaptureID]*model.TaskStatus, error)

	// GetTaskPositions returns the task positions.
	GetTaskPositions() (map[model.CaptureID]*model.TaskPosition, error)
}

// GetTaskStatuses implements InfoProvider for BaseScheduleDispatcher.
func (s *BaseScheduleDispatcher) GetTaskStatuses() (map[model.CaptureID]*model.TaskStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	keyspansPerCapture := s.keyspans.GetAllKeySpansGroupedByCaptures()
	ret := make(map[model.CaptureID]*model.TaskStatus, len(keyspansPerCapture))
	for captureID, keyspans := range keyspansPerCapture {
		ret[captureID] = &model.TaskStatus{
			KeySpans:  make(map[model.KeySpanID]*model.KeySpanReplicaInfo),
			Operation: make(map[model.KeySpanID]*model.KeySpanOperation),
		}
		for keyspanID, record := range keyspans {
			ret[captureID].KeySpans[keyspanID] = &model.KeySpanReplicaInfo{
				StartTs: 0, // We no longer maintain this information
			}
			switch record.Status {
			case util.RunningKeySpan:
				continue
			case util.AddingKeySpan:
				ret[captureID].Operation[keyspanID] = &model.KeySpanOperation{
					Delete:     false,
					Status:     model.OperDispatched,
					BoundaryTs: 0, // We no longer maintain this information
				}
			case util.RemovingKeySpan:
				ret[captureID].Operation[keyspanID] = &model.KeySpanOperation{
					Delete:     true,
					Status:     model.OperDispatched,
					BoundaryTs: 0, // We no longer maintain this information
				}
			}
		}
	}

	return ret, nil
}

// GetTaskPositions implements InfoProvider for BaseScheduleDispatcher.
func (s *BaseScheduleDispatcher) GetTaskPositions() (map[model.CaptureID]*model.TaskPosition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ret := make(map[model.CaptureID]*model.TaskPosition, len(s.captureStatus))
	for captureID, captureStatus := range s.captureStatus {
		ret[captureID] = &model.TaskPosition{
			CheckPointTs: captureStatus.CheckpointTs,
			ResolvedTs:   captureStatus.ResolvedTs,
		}
	}

	return ret, nil
}
