/*
Copyright (C) 2022-2024 ApeCloud Co., Ltd

This file is part of KubeBlocks project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package user

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/apecloud/kubeblocks/pkg/constant"
	"github.com/apecloud/kubeblocks/pkg/kb_agent/actions"
	"github.com/apecloud/kubeblocks/pkg/kb_agent/util"
)

type ListUsers struct {
	actions.Base
}

var listusers actions.Action = &ListUsers{}

func init() {
	err := actions.Register("listusers", listusers)
	if err != nil {
		panic(err.Error())
	}
}

func (s *ListUsers) Init(ctx context.Context) error {
	s.Logger = ctrl.Log.WithName("listusers")
	s.Action = constant.ListUsersAction
	return s.Base.Init(ctx)
}

func (s *ListUsers) IsReadonly(ctx context.Context) bool {
	return true
}

func (s *ListUsers) Do(ctx context.Context, req *actions.OpsRequest) (*actions.OpsResponse, error) {
	resp := actions.NewOpsResponse(util.ListUsersOp)

	result, err := s.Handler.ListUsers(ctx)
	if err != nil {
		s.Logger.Info("executing listusers error", "error", err)
		return resp, err
	}

	resp.Data["users"] = result
	return resp.WithSuccess("")
}
