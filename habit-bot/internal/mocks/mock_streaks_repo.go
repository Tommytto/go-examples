// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Tommytto/habit-bot/internal/repos (interfaces: StreaksRepo)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	time "time"

	repos "github.com/Tommytto/habit-bot/internal/repos"
	gomock "github.com/golang/mock/gomock"
)

// MockStreaksRepo is a mock of StreaksRepo interface.
type MockStreaksRepo struct {
	ctrl     *gomock.Controller
	recorder *MockStreaksRepoMockRecorder
}

// MockStreaksRepoMockRecorder is the mock recorder for MockStreaksRepo.
type MockStreaksRepoMockRecorder struct {
	mock *MockStreaksRepo
}

// NewMockStreaksRepo creates a new mock instance.
func NewMockStreaksRepo(ctrl *gomock.Controller) *MockStreaksRepo {
	mock := &MockStreaksRepo{ctrl: ctrl}
	mock.recorder = &MockStreaksRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStreaksRepo) EXPECT() *MockStreaksRepoMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockStreaksRepo) Create(arg0 *repos.StreakEntity) (*repos.StreakEntity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0)
	ret0, _ := ret[0].(*repos.StreakEntity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockStreaksRepoMockRecorder) Create(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockStreaksRepo)(nil).Create), arg0)
}

// Get mocks base method.
func (m *MockStreaksRepo) Get(arg0 string) (*repos.StreakEntity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*repos.StreakEntity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStreaksRepoMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStreaksRepo)(nil).Get), arg0)
}

// GetAll mocks base method.
func (m *MockStreaksRepo) GetAll(arg0 string) ([]*repos.StreakEntity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", arg0)
	ret0, _ := ret[0].([]*repos.StreakEntity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockStreaksRepoMockRecorder) GetAll(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockStreaksRepo)(nil).GetAll), arg0)
}

// GetCurrentStreak mocks base method.
func (m *MockStreaksRepo) GetCurrentStreak(arg0 string, arg1 time.Time) (*repos.StreakEntity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentStreak", arg0, arg1)
	ret0, _ := ret[0].(*repos.StreakEntity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCurrentStreak indicates an expected call of GetCurrentStreak.
func (mr *MockStreaksRepoMockRecorder) GetCurrentStreak(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentStreak", reflect.TypeOf((*MockStreaksRepo)(nil).GetCurrentStreak), arg0, arg1)
}

// UpdateOne mocks base method.
func (m *MockStreaksRepo) UpdateOne(arg0 string, arg1 time.Time, arg2 map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOne", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOne indicates an expected call of UpdateOne.
func (mr *MockStreaksRepoMockRecorder) UpdateOne(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOne", reflect.TypeOf((*MockStreaksRepo)(nil).UpdateOne), arg0, arg1, arg2)
}
