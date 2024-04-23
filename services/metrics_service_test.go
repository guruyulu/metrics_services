package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/guruyulu/metrics_services/services"
	"github.com/guruyulu/metrics_services/services/mocks"
)

func TestFetchCPUMetrics(t *testing.T) {
	t.Run("CPU metrics data fetched, but DB metrics is not fetched", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		mockAPIClient.On("Query", context.Background(), "my_counter", time.Now()).Return(nil, nil, errors.New("failed to fetch DB metrics"))

		result, err := services.FetchCPUMetrics(mockAPIClient)

		assert.Error(t, err, "Expected error when DB metrics fetch fails")
		assert.Empty(t, result, "Expected empty result when DB metrics fetch fails")
	})

	t.Run("CPU metrics data not fetched, but DB metrics is not fetched", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		mockAPIClient.On("Query", context.Background(), "cpu_query", time.Now()).Return(nil, nil, errors.New("failed to fetch CPU metrics"))

		result, err := services.FetchCPUMetrics(mockAPIClient)

		assert.Error(t, err, "Expected error when both CPU and DB metrics fetch fails")
		assert.Empty(t, result, "Expected empty result when both CPU and DB metrics fetch fails")
	})

	t.Run("CPU metrics data fetched, and DB metrics also fetched", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		mockAPIClient.On("Query", context.Background(), "cpu_query", time.Now()).Return(nil, nil, nil)

		result, err := services.FetchCPUMetrics(mockAPIClient)

		assert.NoError(t, err, "Expected no error when both CPU and DB metrics fetch succeed")
		assert.NotEmpty(t, result, "Expected non-empty result when both CPU and DB metrics fetch succeed")
	})

	t.Run("CPU metrics data not fetched, but DB metrics is fetched", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		mockAPIClient.On("Query", context.Background(), "cpu_query", time.Now()).Return(nil, nil, errors.New("failed to fetch CPU metrics"))

		result, err := services.FetchCPUMetrics(mockAPIClient)

		assert.Error(t, err, "Expected error when CPU metrics fetch fails but DB metrics fetch succeeds")
		assert.Empty(t, result, "Expected empty result when CPU metrics fetch fails but DB metrics fetch succeeds")
	})

	t.Run("Wrong format data received", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		mockAPIClient.On("Query", context.Background(), "cpu_query", time.Now()).Return(nil, nil, errors.New("wrong format data received"))

		result, err := services.FetchCPUMetrics(mockAPIClient)

		assert.Error(t, err, "Expected error when wrong format data received")
		assert.Empty(t, result, "Expected empty result when wrong format data received")
	})

	t.Run("Timeout while fetching data - server error", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		mockAPIClient.On("Query", context.Background(), "cpu_query", time.Now()).Return(nil, nil, errors.New("timeout while fetching data"))

		result, err := services.FetchCPUMetrics(mockAPIClient)

		assert.Error(t, err, "Expected error when server timeout occurs")
		assert.Empty(t, result, "Expected empty result when server timeout occurs")
	})

	t.Run("Nil condition handle", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		mockAPIClient.On("Query", context.Background(), "cpu_query", time.Now()).Return(nil, nil, nil)

		result, err := services.FetchCPUMetrics(mockAPIClient)

		assert.NoError(t, err, "Expected no error when nil condition handled")
		assert.NotEmpty(t, result, "Expected non-empty result when nil condition handled")
	})

	t.Run("Connection refused", func(t *testing.T) {
		mockAPIClient := &mocks.APIClient{}

		mockAPIClient.On("Query", context.Background(), "cpu_query", time.Now()).Return(nil, nil, errors.New("connection refused"))

		result, err := services.FetchCPUMetrics(mockAPIClient)

		assert.Error(t, err, "Expected error when connection refused")
		assert.Empty(t, result, "Expected empty result when connection refused")
	})
}
