// Code generated by mockery v2.53.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/servers/scrapper"
)

// MockNotifierStorage is an autogenerated mock type for the Storage type
type MockNotifierStorage struct {
	mock.Mock
}

type MockNotifierStorage_Expecter struct {
	mock *mock.Mock
}

func (_m *MockNotifierStorage) EXPECT() *MockNotifierStorage_Expecter {
	return &MockNotifierStorage_Expecter{mock: &_m.Mock}
}

// GetChatIDs provides a mock function with given fields: ctx
func (_m *MockNotifierStorage) GetChatIDs(ctx context.Context) ([]int64, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetChatIDs")
	}

	var r0 []int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]int64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []int64); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int64)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockNotifierStorage_GetChatIDs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetChatIDs'
type MockNotifierStorage_GetChatIDs_Call struct {
	*mock.Call
}

// GetChatIDs is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockNotifierStorage_Expecter) GetChatIDs(ctx interface{}) *MockNotifierStorage_GetChatIDs_Call {
	return &MockNotifierStorage_GetChatIDs_Call{Call: _e.mock.On("GetChatIDs", ctx)}
}

func (_c *MockNotifierStorage_GetChatIDs_Call) Run(run func(ctx context.Context)) *MockNotifierStorage_GetChatIDs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockNotifierStorage_GetChatIDs_Call) Return(_a0 []int64, _a1 error) *MockNotifierStorage_GetChatIDs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockNotifierStorage_GetChatIDs_Call) RunAndReturn(run func(context.Context) ([]int64, error)) *MockNotifierStorage_GetChatIDs_Call {
	_c.Call.Return(run)
	return _c
}

// GetLinksWithChatActive provides a mock function with given fields: ctx, chatID
func (_m *MockNotifierStorage) GetLinksWithChatActive(ctx context.Context, chatID int64) ([]scrapperapi.LinkResponse, error) {
	ret := _m.Called(ctx, chatID)

	if len(ret) == 0 {
		panic("no return value specified for GetLinksWithChatActive")
	}

	var r0 []scrapperapi.LinkResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) ([]scrapperapi.LinkResponse, error)); ok {
		return rf(ctx, chatID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) []scrapperapi.LinkResponse); ok {
		r0 = rf(ctx, chatID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]scrapperapi.LinkResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, chatID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockNotifierStorage_GetLinksWithChatActive_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLinksWithChatActive'
type MockNotifierStorage_GetLinksWithChatActive_Call struct {
	*mock.Call
}

// GetLinksWithChatActive is a helper method to define mock.On call
//   - ctx context.Context
//   - chatID int64
func (_e *MockNotifierStorage_Expecter) GetLinksWithChatActive(ctx interface{}, chatID interface{}) *MockNotifierStorage_GetLinksWithChatActive_Call {
	return &MockNotifierStorage_GetLinksWithChatActive_Call{Call: _e.mock.On("GetLinksWithChatActive", ctx, chatID)}
}

func (_c *MockNotifierStorage_GetLinksWithChatActive_Call) Run(run func(ctx context.Context, chatID int64)) *MockNotifierStorage_GetLinksWithChatActive_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *MockNotifierStorage_GetLinksWithChatActive_Call) Return(_a0 []scrapperapi.LinkResponse, _a1 error) *MockNotifierStorage_GetLinksWithChatActive_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockNotifierStorage_GetLinksWithChatActive_Call) RunAndReturn(run func(context.Context, int64) ([]scrapperapi.LinkResponse, error)) *MockNotifierStorage_GetLinksWithChatActive_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateLinkActivity provides a mock function with given fields: ctx, linkID, status
func (_m *MockNotifierStorage) UpdateLinkActivity(ctx context.Context, linkID int64, status bool) error {
	ret := _m.Called(ctx, linkID, status)

	if len(ret) == 0 {
		panic("no return value specified for UpdateLinkActivity")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, bool) error); ok {
		r0 = rf(ctx, linkID, status)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockNotifierStorage_UpdateLinkActivity_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateLinkActivity'
type MockNotifierStorage_UpdateLinkActivity_Call struct {
	*mock.Call
}

// UpdateLinkActivity is a helper method to define mock.On call
//   - ctx context.Context
//   - linkID int64
//   - status bool
func (_e *MockNotifierStorage_Expecter) UpdateLinkActivity(ctx interface{}, linkID interface{}, status interface{}) *MockNotifierStorage_UpdateLinkActivity_Call {
	return &MockNotifierStorage_UpdateLinkActivity_Call{Call: _e.mock.On("UpdateLinkActivity", ctx, linkID, status)}
}

func (_c *MockNotifierStorage_UpdateLinkActivity_Call) Run(run func(ctx context.Context, linkID int64, status bool)) *MockNotifierStorage_UpdateLinkActivity_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64), args[2].(bool))
	})
	return _c
}

func (_c *MockNotifierStorage_UpdateLinkActivity_Call) Return(_a0 error) *MockNotifierStorage_UpdateLinkActivity_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockNotifierStorage_UpdateLinkActivity_Call) RunAndReturn(run func(context.Context, int64, bool) error) *MockNotifierStorage_UpdateLinkActivity_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockNotifierStorage creates a new instance of MockNotifierStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockNotifierStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockNotifierStorage {
	mock := &MockNotifierStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
