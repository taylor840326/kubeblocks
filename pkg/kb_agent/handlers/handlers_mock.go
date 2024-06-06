// /*
// Copyright (C) 2022-2024 ApeCloud Co., Ltd
//
// This file is part of KubeBlocks project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
// */
//
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/apecloud/kubeblocks/pkg/kb_agent/handlers (interfaces: Handler)

// Package handlers is a generated GoMock package.
package handlers

import (
	context "context"
	reflect "reflect"

	models "github.com/apecloud/kubeblocks/pkg/kb_agent/handlers/models"
	logr "github.com/go-logr/logr"
	gomock "github.com/golang/mock/gomock"
)

// MockHandler is a mock of Handler interface.
type MockHandler struct {
	ctrl     *gomock.Controller
	recorder *MockHandlerMockRecorder
}

// MockHandlerMockRecorder is the mock recorder for MockHandler.
type MockHandlerMockRecorder struct {
	mock *MockHandler
}

// NewMockHandler creates a new mock instance.
func NewMockHandler(ctrl *gomock.Controller) *MockHandler {
	mock := &MockHandler{ctrl: ctrl}
	mock.recorder = &MockHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHandler) EXPECT() *MockHandlerMockRecorder {
	return m.recorder
}

// CreateUser mocks base method.
func (m *MockHandler) CreateUser(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockHandlerMockRecorder) CreateUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockHandler)(nil).CreateUser), arg0, arg1, arg2)
}

// DataDump mocks base method.
func (m *MockHandler) DataDump(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DataDump", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DataDump indicates an expected call of DataDump.
func (mr *MockHandlerMockRecorder) DataDump(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DataDump", reflect.TypeOf((*MockHandler)(nil).DataDump), arg0)
}

// DataLoad mocks base method.
func (m *MockHandler) DataLoad(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DataLoad", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DataLoad indicates an expected call of DataLoad.
func (mr *MockHandlerMockRecorder) DataLoad(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DataLoad", reflect.TypeOf((*MockHandler)(nil).DataLoad), arg0)
}

// DeleteUser mocks base method.
func (m *MockHandler) DeleteUser(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockHandlerMockRecorder) DeleteUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockHandler)(nil).DeleteUser), arg0, arg1)
}

// DescribeUser mocks base method.
func (m *MockHandler) DescribeUser(arg0 context.Context, arg1 string) (*models.UserInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeUser", arg0, arg1)
	ret0, _ := ret[0].(*models.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeUser indicates an expected call of DescribeUser.
func (mr *MockHandlerMockRecorder) DescribeUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeUser", reflect.TypeOf((*MockHandler)(nil).DescribeUser), arg0, arg1)
}

// GetCurrentMemberName mocks base method.
func (m *MockHandler) GetCurrentMemberName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentMemberName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetCurrentMemberName indicates an expected call of GetCurrentMemberName.
func (mr *MockHandlerMockRecorder) GetCurrentMemberName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentMemberName", reflect.TypeOf((*MockHandler)(nil).GetCurrentMemberName))
}

// GetLogger mocks base method.
func (m *MockHandler) GetLogger() logr.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLogger")
	ret0, _ := ret[0].(logr.Logger)
	return ret0
}

// GetLogger indicates an expected call of GetLogger.
func (mr *MockHandlerMockRecorder) GetLogger() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLogger", reflect.TypeOf((*MockHandler)(nil).GetLogger))
}

// GetReplicaRole mocks base method.
func (m *MockHandler) GetReplicaRole(arg0 context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReplicaRole", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetReplicaRole indicates an expected call of GetReplicaRole.
func (mr *MockHandlerMockRecorder) GetReplicaRole(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReplicaRole", reflect.TypeOf((*MockHandler)(nil).GetReplicaRole), arg0)
}

// GrantUserRole mocks base method.
func (m *MockHandler) GrantUserRole(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GrantUserRole", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// GrantUserRole indicates an expected call of GrantUserRole.
func (mr *MockHandlerMockRecorder) GrantUserRole(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GrantUserRole", reflect.TypeOf((*MockHandler)(nil).GrantUserRole), arg0, arg1, arg2)
}

// HealthyCheck mocks base method.
func (m *MockHandler) HealthyCheck(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HealthyCheck", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// HealthyCheck indicates an expected call of HealthyCheck.
func (mr *MockHandlerMockRecorder) HealthyCheck(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HealthyCheck", reflect.TypeOf((*MockHandler)(nil).HealthyCheck), arg0)
}

// IsDBStartupReady mocks base method.
func (m *MockHandler) IsDBStartupReady() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDBStartupReady")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsDBStartupReady indicates an expected call of IsDBStartupReady.
func (mr *MockHandlerMockRecorder) IsDBStartupReady() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDBStartupReady", reflect.TypeOf((*MockHandler)(nil).IsDBStartupReady))
}

// IsRunning mocks base method.
func (m *MockHandler) IsRunning() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsRunning")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsRunning indicates an expected call of IsRunning.
func (mr *MockHandlerMockRecorder) IsRunning() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsRunning", reflect.TypeOf((*MockHandler)(nil).IsRunning))
}

// JoinMember mocks base method.
func (m *MockHandler) JoinMember(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "JoinMember", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// JoinMember indicates an expected call of JoinMember.
func (mr *MockHandlerMockRecorder) JoinMember(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JoinMember", reflect.TypeOf((*MockHandler)(nil).JoinMember), arg0, arg1)
}

// LeaveMember mocks base method.
func (m *MockHandler) LeaveMember(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LeaveMember", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// LeaveMember indicates an expected call of LeaveMember.
func (mr *MockHandlerMockRecorder) LeaveMember(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeaveMember", reflect.TypeOf((*MockHandler)(nil).LeaveMember), arg0, arg1)
}

// ListSystemAccounts mocks base method.
func (m *MockHandler) ListSystemAccounts(arg0 context.Context) ([]models.UserInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSystemAccounts", arg0)
	ret0, _ := ret[0].([]models.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSystemAccounts indicates an expected call of ListSystemAccounts.
func (mr *MockHandlerMockRecorder) ListSystemAccounts(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSystemAccounts", reflect.TypeOf((*MockHandler)(nil).ListSystemAccounts), arg0)
}

// ListUsers mocks base method.
func (m *MockHandler) ListUsers(arg0 context.Context) ([]models.UserInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUsers", arg0)
	ret0, _ := ret[0].([]models.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsers indicates an expected call of ListUsers.
func (mr *MockHandlerMockRecorder) ListUsers(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsers", reflect.TypeOf((*MockHandler)(nil).ListUsers), arg0)
}

// PostProvision mocks base method.
func (m *MockHandler) PostProvision(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostProvision", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// PostProvision indicates an expected call of PostProvision.
func (mr *MockHandlerMockRecorder) PostProvision(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostProvision", reflect.TypeOf((*MockHandler)(nil).PostProvision), arg0)
}

// PreTerminate mocks base method.
func (m *MockHandler) PreTerminate(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PreTerminate", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// PreTerminate indicates an expected call of PreTerminate.
func (mr *MockHandlerMockRecorder) PreTerminate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PreTerminate", reflect.TypeOf((*MockHandler)(nil).PreTerminate), arg0)
}

// ReadOnly mocks base method.
func (m *MockHandler) ReadOnly(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadOnly", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReadOnly indicates an expected call of ReadOnly.
func (mr *MockHandlerMockRecorder) ReadOnly(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadOnly", reflect.TypeOf((*MockHandler)(nil).ReadOnly), arg0, arg1)
}

// ReadWrite mocks base method.
func (m *MockHandler) ReadWrite(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadWrite", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReadWrite indicates an expected call of ReadWrite.
func (mr *MockHandlerMockRecorder) ReadWrite(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadWrite", reflect.TypeOf((*MockHandler)(nil).ReadWrite), arg0, arg1)
}

// RevokeUserRole mocks base method.
func (m *MockHandler) RevokeUserRole(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RevokeUserRole", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// RevokeUserRole indicates an expected call of RevokeUserRole.
func (mr *MockHandlerMockRecorder) RevokeUserRole(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RevokeUserRole", reflect.TypeOf((*MockHandler)(nil).RevokeUserRole), arg0, arg1, arg2)
}

// ShutDownWithWait mocks base method.
func (m *MockHandler) ShutDownWithWait() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ShutDownWithWait")
}

// ShutDownWithWait indicates an expected call of ShutDownWithWait.
func (mr *MockHandlerMockRecorder) ShutDownWithWait() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShutDownWithWait", reflect.TypeOf((*MockHandler)(nil).ShutDownWithWait))
}

// Switchover mocks base method.
func (m *MockHandler) Switchover(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Switchover", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Switchover indicates an expected call of Switchover.
func (mr *MockHandlerMockRecorder) Switchover(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Switchover", reflect.TypeOf((*MockHandler)(nil).Switchover), arg0, arg1, arg2)
}
