

package mocks

import (
	context "context"

	model "github.com/prometheus/common/model"
	mock "github.com/stretchr/testify/mock"

	time "time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type APIClient struct {
	mock.Mock
}

func (_m *APIClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, v1.Warnings, error) {
	ret := _m.Called(ctx, query, ts)

	if len(ret) == 0 {
		panic("no return value specified for Query")
	}

	var r0 model.Value
	var r1 v1.Warnings
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time) (model.Value, v1.Warnings, error)); ok {
		return rf(ctx, query, ts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time) model.Value); ok {
		r0 = rf(ctx, query, ts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.Value)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, time.Time) v1.Warnings); ok {
		r1 = rf(ctx, query, ts)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(v1.Warnings)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, time.Time) error); ok {
		r2 = rf(ctx, query, ts)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

func NewAPIClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *APIClient {
	mock := &APIClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
