package ha

import (
	"context"
	"github.com/apecloud/kubeblocks/cmd/probe/internal/component/configuration_store"
	"github.com/dapr/components-contrib/bindings"
)

type DB interface {
	Promote(ctx context.Context, podName string) error
	Demote(ctx context.Context, podName string) error

	GetStatus(ctx context.Context) (string, error)
	GetExtra(ctx context.Context) (map[string]string, error)
	GetOpTime(ctx context.Context) (int64, error)
	IsLeader(ctx context.Context) (bool, error)
	IsHealthiest(ctx context.Context, podName string) bool
	HandleFollow(ctx context.Context, leader *configuration_store.Leader, podName string) error
	EnforcePrimaryRole(ctx context.Context, podName string) error
	ProcessManualSwitchoverFromLeader(ctx context.Context, podName string) (bool, error)
	ProcessManualSwitchoverFromNoLeader(ctx context.Context, podName string) bool
	InitDelay() error
	Init(metadata bindings.Metadata) error
	Follow(ctx context.Context, podName string, needRestart bool, leader string) error

	DbConn
	DbTool
	ProcessControl
}

type DbConn interface {
	GetSysID(ctx context.Context) (string, error)
}

type DbTool interface {
	ExecCmd(ctx context.Context, podName, cmd string) (map[string]string, error)
}

type ProcessControl interface {
	Stop(ctx context.Context) error
}
