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

package util

import (
	"fmt"
	"net/http"
	"strconv"

	"paddleflow/pkg/apiserver/common"
	"paddleflow/pkg/common/logger"
)

const (
	PaddleflowRouterPrefix    = "/api/paddleflow"
	PaddleflowRouterVersionV1 = "/v1"

	DefaultMaxKeys = 50
	ListPageMax    = 1000

	ParamKeyQueueName  = "queueName"
	ParamKeyRunID      = "runID"
	ParamKeyRunCacheID = "runCacheID"
	ParamKeyPipelineID = "pipelineID"

	QueryKeyAction   = "action"
	QueryActionStop  = "stop"
	QueryActionRetry = "retry"
	QueryActionClose = "close"

	QueryKeyMarker  = "marker"
	QueryKeyMaxKeys = "maxKeys"

	QueryKeyUserFilter = "userFilter"
	QueryKeyFsFilter   = "fsFilter"
	QueryKeyNameFilter = "nameFilter"
	QueryKeyRunFilter  = "runFilter"
	QueryKeyTypeFilter = "typeFilter"
	QueryKeyPathFilter = "pathFilter"
	QueryKeyUser       = "user"
	QueryKeyName       = "name"
	QueryKeyUserName   = "username"
	QueryResourceType  = "resourceType"
	QueryResourceID    = "resourceID"

	ParamKeyClusterName   = "clusterName"
	ParamKeyClusterNames  = "clusterNames"
	ParamKeyClusterStatus = "clusterStatus"

	QueryFsPath = "fsPath"
	QueryFsName = "fsName"
	QueryFsname = "fsname"
	QueryPath   = "path"

	// cluster name最大长度
	ClusterNameMaxLength = 255
)

func GetQueryMaxKeys(ctx *logger.RequestContext, r *http.Request) (int, error) {
	var maxKeys int
	queryMaxKeys := r.URL.Query().Get(QueryKeyMaxKeys)
	if queryMaxKeys == "" {
		maxKeys = DefaultMaxKeys
	} else {
		mk, err := strconv.Atoi(queryMaxKeys)
		if err != nil {
			ctx.Logging().Errorf("query maxKeys[%s] is invalid.", queryMaxKeys)
			ctx.ErrorMessage = fmt.Sprintf("query maxKeys[%s] is invalid.", queryMaxKeys)
			ctx.ErrorCode = common.InvalidURI
			return 0, common.InvalidMaxKeysError(queryMaxKeys)
		}
		maxKeys = mk
	}
	return maxKeys, nil
}
