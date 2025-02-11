/*
Copyright (c) 2021 PaddlePaddle Authors. All Rights Reserve.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import "strings"

const (
	StatusQueueCreating = "creating"
	StatusQueueOpen     = "open"
	StatusQueueClosing  = "closing"
	StatusQueueClosed   = "closed"

	StatusRunInitiating  = "initiating"
	StatusRunPending     = "pending"
	StatusRunRunning     = "running"
	StatusRunSucceeded   = "succeeded"
	StatusRunFailed      = "failed"
	StatusRunTerminating = "terminating"
	StatusRunTerminated  = "terminated"

	WfEventKeyRunID   = "runID"
	WfEventKeyStatus  = "status"
	WfEventKeyRuntime = "runtime"
)

var (
	RunFinalStatus = []string{
		StatusRunFailed,
		StatusRunSucceeded,
		StatusRunTerminated,
	}

	RunActiveStatus = []string{
		StatusRunPending,
		StatusRunRunning,
		StatusRunTerminating,
		StatusRunInitiating,
	}
)

func IsRunFinalStatus(status string) bool {
	if strings.EqualFold(status, StatusRunFailed) ||
		strings.EqualFold(status, StatusRunSucceeded) ||
		strings.EqualFold(status, StatusRunTerminated) {
		return true
	}
	return false
}
