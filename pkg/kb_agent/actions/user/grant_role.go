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
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/apecloud/kubeblocks/pkg/constant"
	"github.com/apecloud/kubeblocks/pkg/kb_agent/actions"
	"github.com/apecloud/kubeblocks/pkg/kb_agent/util"
)

type GrantRole struct {
	actions.Base
}

var grantRole actions.Action = &GrantRole{}

func init() {
	err := actions.Register(strings.ToLower(string(util.GrantUserRoleOp)), grantRole)
	if err != nil {
		panic(err.Error())
	}
}

func (s *GrantRole) Init(ctx context.Context) error {
	s.Logger = ctrl.Log.WithName("grantRole")
	s.Action = constant.GrantRoleAction
	return s.Base.Init(ctx)
}

func (s *GrantRole) PreCheck(ctx context.Context, req *actions.OpsRequest) error {
	userInfo, err := UserInfoParser(req)
	if err != nil {
		return err
	}

	return userInfo.UserNameValidator()
}

func (s *GrantRole) Do(ctx context.Context, req *actions.OpsRequest) (*actions.OpsResponse, error) {
	userInfo, _ := UserInfoParser(req)
	resp := actions.NewOpsResponse(util.GrantUserRoleOp)

	err := s.Handler.GrantUserRole(ctx, userInfo.UserName, userInfo.RoleName)
	if err != nil {
		s.Logger.Info("executing grantRole error", "error", err)
		return resp, err
	}

	return resp.WithSuccess("")
}
