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

package volume

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/apecloud/kubeblocks/pkg/constant"
	"github.com/apecloud/kubeblocks/pkg/lorry/operations"
	"github.com/apecloud/kubeblocks/pkg/lorry/util"
)

type Unlock struct {
	operations.Base
	Timeout time.Duration
}

var unlock operations.Operation = &Unlock{}

func init() {
	err := operations.Register(strings.ToLower(string(util.UnlockOperation)), unlock)
	if err != nil {
		panic(err.Error())
	}
}

func (s *Unlock) Init(ctx context.Context) error {
	s.Logger = ctrl.Log.WithName("Unlock")
	s.Action = constant.ReadWriteAction
	return s.Base.Init(ctx)
}

func (s *Unlock) Do(ctx context.Context, req *operations.OpsRequest) (*operations.OpsResponse, error) {
	err := s.DBManager.Unlock(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Unlock DB failed")
	}

	return nil, nil
}
