// Code generated by MockGen. DO NOT EDIT.
// Source: cache.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockCache is a mock of Cache interface.
type MockCache struct {
	ctrl     *gomock.Controller
	recorder *MockCacheMockRecorder
}

// MockCacheMockRecorder is the mock recorder for MockCache.
type MockCacheMockRecorder struct {
	mock *MockCache
}

// NewMockCache creates a new mock instance.
func NewMockCache(ctrl *gomock.Controller) *MockCache {
	mock := &MockCache{ctrl: ctrl}
	mock.recorder = &MockCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCache) EXPECT() *MockCacheMockRecorder {
	return m.recorder
}

// DecrBy mocks base method.
func (m *MockCache) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DecrBy", ctx, key, decrement)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DecrBy indicates an expected call of DecrBy.
func (mr *MockCacheMockRecorder) DecrBy(ctx, key, decrement interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DecrBy", reflect.TypeOf((*MockCache)(nil).DecrBy), ctx, key, decrement)
}

// Del mocks base method.
func (m *MockCache) Del(ctx context.Context, keys ...string) (int64, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range keys {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Del", varargs...)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Del indicates an expected call of Del.
func (mr *MockCacheMockRecorder) Del(ctx interface{}, keys ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, keys...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del", reflect.TypeOf((*MockCache)(nil).Del), varargs...)
}

// Exists mocks base method.
func (m *MockCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range keys {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exists", varargs...)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exists indicates an expected call of Exists.
func (mr *MockCacheMockRecorder) Exists(ctx interface{}, keys ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, keys...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockCache)(nil).Exists), varargs...)
}

// Get mocks base method.
func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, key)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockCacheMockRecorder) Get(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCache)(nil).Get), ctx, key)
}

// IncrBy mocks base method.
func (m *MockCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IncrBy", ctx, key, value)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IncrBy indicates an expected call of IncrBy.
func (mr *MockCacheMockRecorder) IncrBy(ctx, key, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncrBy", reflect.TypeOf((*MockCache)(nil).IncrBy), ctx, key, value)
}

// Set mocks base method.
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, key, value, expiration)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Set indicates an expected call of Set.
func (mr *MockCacheMockRecorder) Set(ctx, key, value, expiration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCache)(nil).Set), ctx, key, value, expiration)
}

// SetNX mocks base method.
func (m *MockCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetNX", ctx, key, value, expiration)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetNX indicates an expected call of SetNX.
func (mr *MockCacheMockRecorder) SetNX(ctx, key, value, expiration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetNX", reflect.TypeOf((*MockCache)(nil).SetNX), ctx, key, value, expiration)
}
