package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/guruyulu/metrics_services/services"
	"github.com/guruyulu/metrics_services/services/mocks"
)

const ScalarType model.ValueType = 0

// MockAPIClient is a mock implementation of the APIClient interface
type MockAPIClient struct {
	mock.Mock
}

// Query mocks the Query method of the APIClient interface
func (m *MockAPIClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, v1.Warnings, error) {
	args := m.Called(ctx, query, ts)
	return args.Get(0).(model.Value), args.Get(1).(v1.Warnings), args.Error(2)
}

// MockScalarValue represents a mock implementation of the model.Value interface for a scalar value.
type MockScalarValue struct {
	Value model.SampleValue
}

// Type returns the type of the mock scalar value.
func (s MockScalarValue) Type() model.ValueType {
	return ScalarType
}

func (s MockScalarValue) String() string {
	return fmt.Sprintf("MockScalarValue: %f", s.Value)
}

func TestFetchCPUMetrics(t *testing.T) {
	ctx := context.Background()
	// ts := time.Now()

	tests := []struct {
		name        string
		mockSetup   func(*MockAPIClient)
		expected    string
		expectErr   bool
		expectedErr string
	}{
		{
			name: "successful query execution",
			mockSetup: func(m *MockAPIClient) {
				v := model.Vector{&model.Sample{Value: model.SampleValue(50)}}
				m.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(v, v1.Warnings{}, nil)

			},
			expected:    "Average CPU Usage: 50.00",
			expectErr:   false,
			expectedErr: "",
		},
		{
			name: "query returns non-vector type",
			mockSetup: func(m *MockAPIClient) {
				v := MockScalarValue{Value: model.SampleValue(0)}
				m.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(v, v1.Warnings{}, fmt.Errorf("unexpected query result type: scalar"))
			},
			expectErr:   true,
			expectedErr: "unexpected query result type: scalar",
		},
		{
			name: "query execution failure",
			mockSetup: func(m *MockAPIClient) {
				m.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(model.Vector{}, v1.Warnings{}, errors.New("query failed"))
			},
			expectErr:   true,
			expectedErr: "query failed",
		},
		{
			name: "query returns warnings",
			mockSetup: func(m *MockAPIClient) {
				v := model.Vector{&model.Sample{Value: model.SampleValue(20)}}
				m.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(v, v1.Warnings{"low data availability"}, nil)
			},
			expected:    "Average CPU Usage: 20.00",
			expectErr:   false,
			expectedErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := new(MockAPIClient)
			tc.mockSetup(client)
			result, err := services.FetchCPUMetrics("job_name", "namespace", client)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
			client.AssertExpectations(t)
		})
	}
}

func TestFetchDBConnections(t *testing.T) {
	t.Run("DB metrics data fetched successfully", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		jobName := "hello-app"
		namespace := "hello-app-namespace"

		mockAPIClient.On("Query", context.Background(), "database_connections{instance=\"hello-app.hello-app-namespace.svc.cluster.local:80\", job=\"hello-app\"}", time.Now()).Return(nil, nil, nil)

		result, err := services.FetchDBConnections(jobName, namespace, mockAPIClient)

		assert.NoError(t, err, "Expected no error when DB metrics fetch succeeds")
		assert.NotEmpty(t, result, "Expected non-empty result when DB metrics fetch succeeds")
	})

	t.Run("DB metrics data not fetched", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		jobName := "hello-app"
		namespace := "hello-app-namespace"

		mockAPIClient.On("Query", context.Background(), "database_connections{instance=\"hello-app.hello-app-namespace.svc.cluster.local:80\", job=\"hello-app\"}", time.Now()).Return(nil, nil, errors.New("failed to fetch DB metrics"))

		result, err := services.FetchDBConnections(jobName, namespace, mockAPIClient)

		assert.Error(t, err, "Expected error when DB metrics fetch fails")
		assert.Empty(t, result, "Expected empty result when DB metrics fetch fails")
	})
}
