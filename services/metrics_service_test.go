package services_test

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/guruyulu/metrics_services/services"
// 	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
// 	"github.com/prometheus/common/model"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// const ScalarType model.ValueType = 0

// func TestFetchCPUMetrics(t *testing.T) {
// 	ctx := context.Background()
// 	// ts := time.Now()

// 	tests := []struct {
// 		name        string
// 		mockSetup   func(*MockAPIClient)
// 		expected    string
// 		expectErr   bool
// 		expectedErr string
// 	}{
// 		{
// 			name: "successful query execution",
// 			mockSetup: func(m *MockAPIClient) {
// 				v := model.Vector{&model.Sample{Value: model.SampleValue(50)}}
// 				m.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(v, v1.Warnings{}, nil)

// 			},
// 			expected:    "Average CPU Usage: 50.00",
// 			expectErr:   false,
// 			expectedErr: "",
// 		},
// 		{
// 			name: "query returns non-vector type",
// 			mockSetup: func(m *MockAPIClient) {
// 				v := MockScalarValue{Value: model.SampleValue(0)}
// 				m.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(v, v1.Warnings{}, fmt.Errorf("unexpected query result type: scalar"))
// 			},
// 			expectErr:   true,
// 			expectedErr: "unexpected query result type: scalar",
// 		},
// 		{
// 			name: "query execution failure",
// 			mockSetup: func(m *MockAPIClient) {
// 				m.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(model.Vector{}, v1.Warnings{}, errors.New("query failed"))
// 			},
// 			expectErr:   true,
// 			expectedErr: "query failed",
// 		},
// 		{
// 			name: "query returns warnings",
// 			mockSetup: func(m *MockAPIClient) {
// 				v := model.Vector{&model.Sample{Value: model.SampleValue(20)}}
// 				m.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(v, v1.Warnings{"low data availability"}, nil)
// 			},
// 			expected:    "Average CPU Usage: 20.00",
// 			expectErr:   false,
// 			expectedErr: "",
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			client := new(MockAPIClient)
// 			tc.mockSetup(client)
// 			result, err := services.FetchCPUMetrics("job_name", "namespace", client)
// 			if tc.expectErr {
// 				assert.Error(t, err)
// 				assert.Contains(t, err.Error(), tc.expectedErr)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tc.expected, result)
// 			}
// 			client.AssertExpectations(t)
// 		})
// 	}
// }

// type MockAPIClient struct {
// 	mock.Mock
// }

// func (m *MockAPIClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, v1.Warnings, error) {
// 	args := m.Called(ctx, query, ts)
// 	return args.Get(0).(model.Value), args.Get(1).(v1.Warnings), args.Error(2)
// }

// type MockScalarValue struct {
// 	Value model.SampleValue
// }

// func (s MockScalarValue) Type() model.ValueType {
// 	return ScalarType
// }

// func (s MockScalarValue) String() string {
// 	return fmt.Sprintf("MockScalarValue: %f", s.Value)
// }

// func TestFetchDBConnections(t *testing.T) {
// 	ctx := context.Background()
// 	tests := []struct {
// 		name           string
// 		mockSetup      func(*MockAPIClient)
// 		expectedResult string
// 		expectErr      bool
// 		expectedErr    string
// 	}{
// 		{
// 			name: "successful query execution",
// 			mockSetup: func(m *MockAPIClient) {
// 				v := model.Vector{&model.Sample{Value: model.SampleValue(50)}}
// 				m.On("Query", ctx, "database_connections{instance=\"hello-app.hello-app-namespace.svc.cluster.local:80\", job=\"hello-app\"}", mock.AnythingOfType("time.Time")).Return(v, v1.Warnings{}, nil)
// 			},
// 			expectedResult: "Average DB Usage: 50.00",
// 			expectErr:      false,
// 			expectedErr:    "",
// 		},
// 		{
// 			name: "query returns non-vector type",
// 			mockSetup: func(m *MockAPIClient) {
// 				v := MockScalarValue{Value: model.SampleValue(0)}
// 				m.On("Query", ctx, "database_connections{instance=\"hello-app.hello-app-namespace.svc.cluster.local:80\", job=\"hello-app\"}", mock.AnythingOfType("time.Time")).Return(v, v1.Warnings{}, fmt.Errorf("unexpected query result type: scalar"))
// 			},
// 			expectErr:   true,
// 			expectedErr: "unexpected query result type: scalar",
// 		},
// 		{
// 			name: "query fails",
// 			mockSetup: func(m *MockAPIClient) {
// 				m.On("Query", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(model.Vector{}, v1.Warnings{}, fmt.Errorf("query failed"))
// 			},
// 			expectErr:   true,
// 			expectedErr: "query failed",
// 		},
// 		{
// 			name: "query returns warnings",
// 			mockSetup: func(m *MockAPIClient) {
// 				v := model.Vector{&model.Sample{Value: model.SampleValue(50)}}
// 				warnings := v1.Warnings{"Database Connections Exceeded", "Database Connections Under Utilize"}
// 				m.On("Query", ctx, "database_connections{instance=\"hello-app.hello-app-namespace.svc.cluster.local:80\", job=\"hello-app\"}", mock.AnythingOfType("time.Time")).Return(v, warnings, nil)
// 			},
// 			expectedResult: "Average DB Usage: 50.00",
// 			expectErr:      false,
// 			expectedErr:    "",
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			client := new(MockAPIClient)
// 			tc.mockSetup(client)
// 			result, err := services.FetchDBConnections("hello-app", "hello-app-namespace", client)
// 			if err != nil {
// 				if !tc.expectErr {
// 					t.Errorf("unexpected error: %v", err)
// 				} else if tc.expectedErr != "" && err.Error() != tc.expectedErr {
// 					t.Errorf("expected error message '%s', got '%v'", tc.expectedErr, err)
// 					fmt.Println("Actual error:", err.Error())
// 				}
// 				return
// 			}
// 			if tc.expectErr {
// 				t.Error("expected an error, but got none")
// 				return
// 			}
// 			assert.Equal(t, tc.expectedResult, result, "result mismatch")
// 			client.AssertExpectations(t)
// 		})
// 	}
// }
