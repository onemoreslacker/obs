// Code generated by mockery v2.53.2. DO NOT EDIT.

package mocks

import (
	context "context"
	http "net/http"

	mock "github.com/stretchr/testify/mock"

	scrapperclient "github.com/es-debug/backend-academy-2024-go-template/internal/api/openapi/v1/clients/scrapper"
)

// MockScrapperClient is an autogenerated mock type for the ScrapperClient type
type MockScrapperClient struct {
	mock.Mock
}

type MockScrapperClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockScrapperClient) EXPECT() *MockScrapperClient_Expecter {
	return &MockScrapperClient_Expecter{mock: &_m.Mock}
}

// DeleteLinks provides a mock function with given fields: ctx, params, body, reqEditors
func (_m *MockScrapperClient) DeleteLinks(ctx context.Context, params *scrapperclient.DeleteLinksParams, body scrapperclient.RemoveLinkRequest, reqEditors ...scrapperclient.RequestEditorFn) (*http.Response, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params, body)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteLinks")
	}

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *scrapperclient.DeleteLinksParams, scrapperclient.RemoveLinkRequest, ...scrapperclient.RequestEditorFn) (*http.Response, error)); ok {
		return rf(ctx, params, body, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *scrapperclient.DeleteLinksParams, scrapperclient.RemoveLinkRequest, ...scrapperclient.RequestEditorFn) *http.Response); ok {
		r0 = rf(ctx, params, body, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *scrapperclient.DeleteLinksParams, scrapperclient.RemoveLinkRequest, ...scrapperclient.RequestEditorFn) error); ok {
		r1 = rf(ctx, params, body, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockScrapperClient_DeleteLinks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteLinks'
type MockScrapperClient_DeleteLinks_Call struct {
	*mock.Call
}

// DeleteLinks is a helper method to define mock.On call
//   - ctx context.Context
//   - params *scrapperclient.DeleteLinksParams
//   - body scrapperclient.RemoveLinkRequest
//   - reqEditors ...scrapperclient.RequestEditorFn
func (_e *MockScrapperClient_Expecter) DeleteLinks(ctx interface{}, params interface{}, body interface{}, reqEditors ...interface{}) *MockScrapperClient_DeleteLinks_Call {
	return &MockScrapperClient_DeleteLinks_Call{Call: _e.mock.On("DeleteLinks",
		append([]interface{}{ctx, params, body}, reqEditors...)...)}
}

func (_c *MockScrapperClient_DeleteLinks_Call) Run(run func(ctx context.Context, params *scrapperclient.DeleteLinksParams, body scrapperclient.RemoveLinkRequest, reqEditors ...scrapperclient.RequestEditorFn)) *MockScrapperClient_DeleteLinks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]scrapperclient.RequestEditorFn, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(scrapperclient.RequestEditorFn)
			}
		}
		run(args[0].(context.Context), args[1].(*scrapperclient.DeleteLinksParams), args[2].(scrapperclient.RemoveLinkRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockScrapperClient_DeleteLinks_Call) Return(_a0 *http.Response, _a1 error) *MockScrapperClient_DeleteLinks_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockScrapperClient_DeleteLinks_Call) RunAndReturn(run func(context.Context, *scrapperclient.DeleteLinksParams, scrapperclient.RemoveLinkRequest, ...scrapperclient.RequestEditorFn) (*http.Response, error)) *MockScrapperClient_DeleteLinks_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteTgChatId provides a mock function with given fields: ctx, id, reqEditors
func (_m *MockScrapperClient) DeleteTgChatId(ctx context.Context, id int64, reqEditors ...scrapperclient.RequestEditorFn) (*http.Response, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteTgChatId")
	}

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...scrapperclient.RequestEditorFn) (*http.Response, error)); ok {
		return rf(ctx, id, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...scrapperclient.RequestEditorFn) *http.Response); ok {
		r0 = rf(ctx, id, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, ...scrapperclient.RequestEditorFn) error); ok {
		r1 = rf(ctx, id, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockScrapperClient_DeleteTgChatId_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteTgChatId'
type MockScrapperClient_DeleteTgChatId_Call struct {
	*mock.Call
}

// DeleteTgChatId is a helper method to define mock.On call
//   - ctx context.Context
//   - id int64
//   - reqEditors ...scrapperclient.RequestEditorFn
func (_e *MockScrapperClient_Expecter) DeleteTgChatId(ctx interface{}, id interface{}, reqEditors ...interface{}) *MockScrapperClient_DeleteTgChatId_Call {
	return &MockScrapperClient_DeleteTgChatId_Call{Call: _e.mock.On("DeleteTgChatId",
		append([]interface{}{ctx, id}, reqEditors...)...)}
}

func (_c *MockScrapperClient_DeleteTgChatId_Call) Run(run func(ctx context.Context, id int64, reqEditors ...scrapperclient.RequestEditorFn)) *MockScrapperClient_DeleteTgChatId_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]scrapperclient.RequestEditorFn, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(scrapperclient.RequestEditorFn)
			}
		}
		run(args[0].(context.Context), args[1].(int64), variadicArgs...)
	})
	return _c
}

func (_c *MockScrapperClient_DeleteTgChatId_Call) Return(_a0 *http.Response, _a1 error) *MockScrapperClient_DeleteTgChatId_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockScrapperClient_DeleteTgChatId_Call) RunAndReturn(run func(context.Context, int64, ...scrapperclient.RequestEditorFn) (*http.Response, error)) *MockScrapperClient_DeleteTgChatId_Call {
	_c.Call.Return(run)
	return _c
}

// GetLinks provides a mock function with given fields: ctx, params, reqEditors
func (_m *MockScrapperClient) GetLinks(ctx context.Context, params *scrapperclient.GetLinksParams, reqEditors ...scrapperclient.RequestEditorFn) (*http.Response, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetLinks")
	}

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *scrapperclient.GetLinksParams, ...scrapperclient.RequestEditorFn) (*http.Response, error)); ok {
		return rf(ctx, params, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *scrapperclient.GetLinksParams, ...scrapperclient.RequestEditorFn) *http.Response); ok {
		r0 = rf(ctx, params, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *scrapperclient.GetLinksParams, ...scrapperclient.RequestEditorFn) error); ok {
		r1 = rf(ctx, params, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockScrapperClient_GetLinks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLinks'
type MockScrapperClient_GetLinks_Call struct {
	*mock.Call
}

// GetLinks is a helper method to define mock.On call
//   - ctx context.Context
//   - params *scrapperclient.GetLinksParams
//   - reqEditors ...scrapperclient.RequestEditorFn
func (_e *MockScrapperClient_Expecter) GetLinks(ctx interface{}, params interface{}, reqEditors ...interface{}) *MockScrapperClient_GetLinks_Call {
	return &MockScrapperClient_GetLinks_Call{Call: _e.mock.On("GetLinks",
		append([]interface{}{ctx, params}, reqEditors...)...)}
}

func (_c *MockScrapperClient_GetLinks_Call) Run(run func(ctx context.Context, params *scrapperclient.GetLinksParams, reqEditors ...scrapperclient.RequestEditorFn)) *MockScrapperClient_GetLinks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]scrapperclient.RequestEditorFn, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(scrapperclient.RequestEditorFn)
			}
		}
		run(args[0].(context.Context), args[1].(*scrapperclient.GetLinksParams), variadicArgs...)
	})
	return _c
}

func (_c *MockScrapperClient_GetLinks_Call) Return(_a0 *http.Response, _a1 error) *MockScrapperClient_GetLinks_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockScrapperClient_GetLinks_Call) RunAndReturn(run func(context.Context, *scrapperclient.GetLinksParams, ...scrapperclient.RequestEditorFn) (*http.Response, error)) *MockScrapperClient_GetLinks_Call {
	_c.Call.Return(run)
	return _c
}

// GetTgChatId provides a mock function with given fields: ctx, id, reqEditors
func (_m *MockScrapperClient) GetTgChatId(ctx context.Context, id int64, reqEditors ...scrapperclient.RequestEditorFn) (*http.Response, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetTgChatId")
	}

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...scrapperclient.RequestEditorFn) (*http.Response, error)); ok {
		return rf(ctx, id, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...scrapperclient.RequestEditorFn) *http.Response); ok {
		r0 = rf(ctx, id, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, ...scrapperclient.RequestEditorFn) error); ok {
		r1 = rf(ctx, id, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockScrapperClient_GetTgChatId_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTgChatId'
type MockScrapperClient_GetTgChatId_Call struct {
	*mock.Call
}

// GetTgChatId is a helper method to define mock.On call
//   - ctx context.Context
//   - id int64
//   - reqEditors ...scrapperclient.RequestEditorFn
func (_e *MockScrapperClient_Expecter) GetTgChatId(ctx interface{}, id interface{}, reqEditors ...interface{}) *MockScrapperClient_GetTgChatId_Call {
	return &MockScrapperClient_GetTgChatId_Call{Call: _e.mock.On("GetTgChatId",
		append([]interface{}{ctx, id}, reqEditors...)...)}
}

func (_c *MockScrapperClient_GetTgChatId_Call) Run(run func(ctx context.Context, id int64, reqEditors ...scrapperclient.RequestEditorFn)) *MockScrapperClient_GetTgChatId_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]scrapperclient.RequestEditorFn, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(scrapperclient.RequestEditorFn)
			}
		}
		run(args[0].(context.Context), args[1].(int64), variadicArgs...)
	})
	return _c
}

func (_c *MockScrapperClient_GetTgChatId_Call) Return(_a0 *http.Response, _a1 error) *MockScrapperClient_GetTgChatId_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockScrapperClient_GetTgChatId_Call) RunAndReturn(run func(context.Context, int64, ...scrapperclient.RequestEditorFn) (*http.Response, error)) *MockScrapperClient_GetTgChatId_Call {
	_c.Call.Return(run)
	return _c
}

// PostLinks provides a mock function with given fields: ctx, params, body, reqEditors
func (_m *MockScrapperClient) PostLinks(ctx context.Context, params *scrapperclient.PostLinksParams, body scrapperclient.AddLinkRequest, reqEditors ...scrapperclient.RequestEditorFn) (*http.Response, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params, body)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for PostLinks")
	}

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *scrapperclient.PostLinksParams, scrapperclient.AddLinkRequest, ...scrapperclient.RequestEditorFn) (*http.Response, error)); ok {
		return rf(ctx, params, body, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *scrapperclient.PostLinksParams, scrapperclient.AddLinkRequest, ...scrapperclient.RequestEditorFn) *http.Response); ok {
		r0 = rf(ctx, params, body, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *scrapperclient.PostLinksParams, scrapperclient.AddLinkRequest, ...scrapperclient.RequestEditorFn) error); ok {
		r1 = rf(ctx, params, body, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockScrapperClient_PostLinks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PostLinks'
type MockScrapperClient_PostLinks_Call struct {
	*mock.Call
}

// PostLinks is a helper method to define mock.On call
//   - ctx context.Context
//   - params *scrapperclient.PostLinksParams
//   - body scrapperclient.AddLinkRequest
//   - reqEditors ...scrapperclient.RequestEditorFn
func (_e *MockScrapperClient_Expecter) PostLinks(ctx interface{}, params interface{}, body interface{}, reqEditors ...interface{}) *MockScrapperClient_PostLinks_Call {
	return &MockScrapperClient_PostLinks_Call{Call: _e.mock.On("PostLinks",
		append([]interface{}{ctx, params, body}, reqEditors...)...)}
}

func (_c *MockScrapperClient_PostLinks_Call) Run(run func(ctx context.Context, params *scrapperclient.PostLinksParams, body scrapperclient.AddLinkRequest, reqEditors ...scrapperclient.RequestEditorFn)) *MockScrapperClient_PostLinks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]scrapperclient.RequestEditorFn, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(scrapperclient.RequestEditorFn)
			}
		}
		run(args[0].(context.Context), args[1].(*scrapperclient.PostLinksParams), args[2].(scrapperclient.AddLinkRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockScrapperClient_PostLinks_Call) Return(_a0 *http.Response, _a1 error) *MockScrapperClient_PostLinks_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockScrapperClient_PostLinks_Call) RunAndReturn(run func(context.Context, *scrapperclient.PostLinksParams, scrapperclient.AddLinkRequest, ...scrapperclient.RequestEditorFn) (*http.Response, error)) *MockScrapperClient_PostLinks_Call {
	_c.Call.Return(run)
	return _c
}

// PostTgChatId provides a mock function with given fields: ctx, id, reqEditors
func (_m *MockScrapperClient) PostTgChatId(ctx context.Context, id int64, reqEditors ...scrapperclient.RequestEditorFn) (*http.Response, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for PostTgChatId")
	}

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...scrapperclient.RequestEditorFn) (*http.Response, error)); ok {
		return rf(ctx, id, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...scrapperclient.RequestEditorFn) *http.Response); ok {
		r0 = rf(ctx, id, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, ...scrapperclient.RequestEditorFn) error); ok {
		r1 = rf(ctx, id, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockScrapperClient_PostTgChatId_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PostTgChatId'
type MockScrapperClient_PostTgChatId_Call struct {
	*mock.Call
}

// PostTgChatId is a helper method to define mock.On call
//   - ctx context.Context
//   - id int64
//   - reqEditors ...scrapperclient.RequestEditorFn
func (_e *MockScrapperClient_Expecter) PostTgChatId(ctx interface{}, id interface{}, reqEditors ...interface{}) *MockScrapperClient_PostTgChatId_Call {
	return &MockScrapperClient_PostTgChatId_Call{Call: _e.mock.On("PostTgChatId",
		append([]interface{}{ctx, id}, reqEditors...)...)}
}

func (_c *MockScrapperClient_PostTgChatId_Call) Run(run func(ctx context.Context, id int64, reqEditors ...scrapperclient.RequestEditorFn)) *MockScrapperClient_PostTgChatId_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]scrapperclient.RequestEditorFn, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(scrapperclient.RequestEditorFn)
			}
		}
		run(args[0].(context.Context), args[1].(int64), variadicArgs...)
	})
	return _c
}

func (_c *MockScrapperClient_PostTgChatId_Call) Return(_a0 *http.Response, _a1 error) *MockScrapperClient_PostTgChatId_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockScrapperClient_PostTgChatId_Call) RunAndReturn(run func(context.Context, int64, ...scrapperclient.RequestEditorFn) (*http.Response, error)) *MockScrapperClient_PostTgChatId_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockScrapperClient creates a new instance of MockScrapperClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockScrapperClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockScrapperClient {
	mock := &MockScrapperClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
