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

package run

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"paddleflow/pkg/apiserver/common"
	"paddleflow/pkg/apiserver/models"
	"paddleflow/pkg/common/database/db_fake"
	"paddleflow/pkg/common/logger"
	"paddleflow/pkg/common/schema"
	"paddleflow/pkg/pipeline"
)

const (
	MockRunID1   = "run-id_1"
	MockRunName1 = "run-name_1"
	MockRootUser = "root"
	MockFsID1    = "fs-id_1"
	MockRunID3   = "run-id_3"

	MockRunID2   = "run-id_2"
	MockRunName2 = "run-name_2"
	MockUserID2  = "user-id_2"
	MockFsID2    = "fs-id_2"
)

func getMockRun1() models.Run {
	run1 := models.Run{
		ID:       MockRunID1,
		Name:     MockRunName1,
		UserName: MockRootUser,
		FsID:     MockFsID1,
		Status:   common.StatusRunPending,
	}
	return run1
}

func getMockRun1_3() models.Run {
	run1 := models.Run{
		ID:         MockRunID3,
		Name:       "run_with_runtime",
		UserName:   MockRootUser,
		FsID:       MockFsID1,
		Status:     common.StatusRunRunning,
		RuntimeRaw: "{\"main\":{\"Id\":\"job-48167450\",\"StepName\":\"run-1862d0fe-main\",\"Command\":\"echo 123; sleep 10\",\"Parameters\":{\"p1\":\"123\"},\"Env\":{\"PF_FS_ID\":\"fs-f899dc604904a17c\",\"PF_JOB_FLAVOUR\":\"flavour-cpu\",\"PF_JOB_MODE\":\"Pod\",\"PF_JOB_NAMESPACE\":\"default\",\"PF_JOB_PVC_NAME\":\"pfs-fs-f899dc604904a17c-pvc\",\"PF_JOB_QUEUE_NAME\":\"queue-3cd24f00\",\"PF_JOB_TYPE\":\"vcjob\",\"PF_RUN_ID\":\"run-1862d0fe\",\"PF_STEP_NAME\":\"main\",\"PF_USER_ID\":\"admin\"},\"StartTime\":\"2021-08-27 17:48:33\",\"EndTime\":\"\",\"Status\":\"running\"}}",
	}
	return run1
}

func getMockRun2() models.Run {
	run2 := models.Run{
		ID:       MockRunID2,
		Name:     MockRunName2,
		UserName: MockUserID2,
		FsID:     MockFsID2,
		Status:   common.StatusRunPending,
	}
	return run2
}

func TestListRunSuccess(t *testing.T) {
	db_fake.InitFakeDB()
	ctx1 := &logger.RequestContext{UserName: MockRootUser}
	ctx2 := &logger.RequestContext{UserName: MockUserID2}
	var err error
	run1 := getMockRun1()
	run1.ID, err = models.CreateRun(ctx1.Logging(), &run1)
	assert.Nil(t, err)
	run2 := getMockRun2()
	run2.ID, err = models.CreateRun(ctx2.Logging(), &run2)
	assert.Nil(t, err)
	run3UnderUser1 := getMockRun1_3()
	run3UnderUser1.ID, err = models.CreateRun(ctx1.Logging(), &run3UnderUser1)
	assert.Nil(t, err)

	emptyFilter := make([]string, 0)
	// test list runs under user1
	listRunResponse, err := ListRun(ctx1, "", 50, emptyFilter, emptyFilter, emptyFilter, emptyFilter)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(listRunResponse.RunList))
	assert.Equal(t, MockRootUser, listRunResponse.RunList[0].UserName)

	// test list runs under user2
	listRunResponse, err = ListRun(ctx2, "", 50, emptyFilter, emptyFilter, emptyFilter, emptyFilter)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(listRunResponse.RunList))
	assert.Equal(t, MockUserID2, listRunResponse.RunList[0].UserName)
}

func TestGetRunSuccess(t *testing.T) {
	db_fake.InitFakeDB()
	ctx := &logger.RequestContext{UserName: MockRootUser}
	var err error
	// test no runtime
	run1 := getMockRun1()
	run1.ID, err = models.CreateRun(ctx.Logging(), &run1)
	assert.Nil(t, err)

	runRsp, err := GetRunByID(ctx, run1.ID)
	assert.Nil(t, err)
	assert.Equal(t, run1.ID, runRsp.ID)
	assert.Equal(t, run1.Name, runRsp.Name)
	assert.Equal(t, run1.Status, runRsp.Status)

	// test has runtime
	run3 := getMockRun1_3()
	run3.ID, err = models.CreateRun(ctx.Logging(), &run3)
	assert.Nil(t, err)
	runRsp, err = GetRunByID(ctx, run3.ID)
	assert.Nil(t, err)
	assert.Equal(t, run3.ID, runRsp.ID)
	assert.Equal(t, run3.Name, runRsp.Name)
	assert.Equal(t, run3.Status, runRsp.Status)
}

func TestGetRunFail(t *testing.T) {
	db_fake.InitFakeDB()
	var err error
	ctx := &logger.RequestContext{UserName: MockRootUser}
	run1 := getMockRun1()
	run1.ID, err = models.CreateRun(ctx.Logging(), &run1)

	// test non-admin user no access to other users' run
	ctxOtherNonAdmin := &logger.RequestContext{UserName: "non-admin"}
	_, err = GetRunByID(ctxOtherNonAdmin, run1.ID)
	assert.NotNil(t, err)
	assert.Equal(t, common.AccessDenied, ctxOtherNonAdmin.ErrorCode)
	assert.Equal(t, common.NoAccessError("non-admin", common.ResourceTypeRun, run1.ID).Error(), err.Error())

	// test no record
	_, err = GetRunByID(ctx, "run-id_non_existed")
	assert.NotNil(t, err)
	assert.Equal(t, common.RunNotFound, ctx.ErrorCode)
	assert.Equal(t, common.NotFoundError(common.ResourceTypeRun, "run-id_non_existed").Error(), err.Error())
}

func TestCallback(t *testing.T) {
	db_fake.InitFakeDB()
	var err error
	ctx := &logger.RequestContext{UserName: MockRootUser}

	// test update activated_at
	run1 := getMockRun1()
	run1.ID, err = models.CreateRun(ctx.Logging(), &run1)
	assert.Nil(t, err)
	runtimeView := schema.RuntimeView{
		"data_preprocess": schema.JobView{
			JobID:  "data_preprocess",
			Status: schema.StatusJobSucceeded,
		},
		"main": {
			JobID:  "data_preprocess",
			Status: schema.StatusJobRunning,
		},
	}
	event1 := pipeline.WorkflowEvent{
		Event: pipeline.WfEventRunUpdate,
		Extra: map[string]interface{}{
			common.WfEventKeyRunID:   run1.ID,
			common.WfEventKeyStatus:  common.StatusRunRunning,
			common.WfEventKeyRuntime: runtimeView,
		},
	}
	f := UpdateRunFunc
	f(run1.ID, &event1)
	updatedRun, err := models.GetRunByID(ctx.Logging(), run1.ID)
	assert.Nil(t, err)
	assert.NotEmpty(t, updatedRun.UpdateTime)
	assert.Equal(t, common.StatusRunRunning, updatedRun.Status)

	// test not update activated_at
	run3 := getMockRun1_3()
	run3.ID, err = models.CreateRun(ctx.Logging(), &run3)
	assert.Nil(t, err)
	f(run3.ID, &event1)
	updatedRun, err = models.GetRunByID(ctx.Logging(), run3.ID)
	assert.Nil(t, err)
	assert.False(t, updatedRun.ActivatedAt.Valid)
	assert.Empty(t, updatedRun.ActivateTime)
}