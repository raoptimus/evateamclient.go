/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMetrics_Register_Success_NoError(t *testing.T) {
	m := NewPrometheusMetrics()
	reg := prometheus.NewRegistry()

	err := m.Register(reg)

	require.NoError(t, err)
}

func TestPrometheusMetrics_Register_AlreadyRegistered_NoError(t *testing.T) {
	m := NewPrometheusMetrics()
	reg := prometheus.NewRegistry()

	err := m.Register(reg)
	require.NoError(t, err)

	err = m.Register(reg)

	require.NoError(t, err)
}

func TestPrometheusMetrics_Unregister_AfterRegister_ReturnsTrue(t *testing.T) {
	m := NewPrometheusMetrics()
	reg := prometheus.NewRegistry()

	err := m.Register(reg)
	require.NoError(t, err)

	ok := m.Unregister(reg)

	assert.True(t, ok)
}

func TestPrometheusMetrics_Unregister_WithoutRegister_ReturnsFalse(t *testing.T) {
	m := NewPrometheusMetrics()
	reg := prometheus.NewRegistry()

	ok := m.Unregister(reg)

	assert.False(t, ok)
}

func TestPrometheusMetrics_RecordRequestDuration_StoresObservation_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		method   string
		host     string
		function string
		duration float64
	}{
		{
			name:     "successful request",
			status:   200,
			method:   "CmfTask.list",
			host:     "api.eva.team",
			function: "TasksList",
			duration: 0.123,
		},
		{
			name:     "error request",
			status:   500,
			method:   "CmfProject.get",
			host:     "api.eva.team",
			function: "ProjectGet",
			duration: 2.5,
		},
		{
			name:     "zero duration",
			status:   200,
			method:   "CmfTag.list",
			host:     "api.eva.team",
			function: "TagList",
			duration: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPrometheusMetrics()
			reg := prometheus.NewRegistry()
			err := m.Register(reg)
			require.NoError(t, err)

			m.RecordRequestDuration(tt.status, tt.method, tt.host, tt.function, tt.duration)

			gathered, err := reg.Gather()
			require.NoError(t, err)
			require.Len(t, gathered, 1)

			mf := gathered[0]
			assert.Equal(t, "eva_client_request_duration_seconds", mf.GetName())
			require.Len(t, mf.GetMetric(), 1)

			metric := mf.GetMetric()[0]
			assert.Equal(t, uint64(1), metric.GetHistogram().GetSampleCount())

			labels := make(map[string]string)
			for _, lp := range metric.GetLabel() {
				labels[lp.GetName()] = lp.GetValue()
			}
			assert.Equal(t, tt.method, labels["method"])
			assert.Equal(t, tt.host, labels["host"])
			assert.Equal(t, tt.function, labels["function"])
		})
	}
}

func TestPrometheusMetrics_RecordRequestDuration_MultipleObservations_AccumulatesCount(t *testing.T) {
	m := NewPrometheusMetrics()
	reg := prometheus.NewRegistry()
	err := m.Register(reg)
	require.NoError(t, err)

	m.RecordRequestDuration(200, "CmfTask.list", "api.eva.team", "TasksList", 0.05)
	m.RecordRequestDuration(200, "CmfTask.list", "api.eva.team", "TasksList", 0.15)
	m.RecordRequestDuration(200, "CmfTask.list", "api.eva.team", "TasksList", 1.0)

	gathered, err := reg.Gather()
	require.NoError(t, err)
	require.Len(t, gathered, 1)

	metric := gathered[0].GetMetric()[0]
	assert.Equal(t, uint64(3), metric.GetHistogram().GetSampleCount())
}

func TestPrometheusMetrics_RecordRequestDuration_DifferentLabels_CreatesDistinctSeries(t *testing.T) {
	m := NewPrometheusMetrics()
	reg := prometheus.NewRegistry()
	err := m.Register(reg)
	require.NoError(t, err)

	m.RecordRequestDuration(200, "CmfTask.list", "api.eva.team", "TasksList", 0.1)
	m.RecordRequestDuration(404, "CmfProject.get", "api.eva.team", "ProjectGet", 0.2)

	gathered, err := reg.Gather()
	require.NoError(t, err)
	require.Len(t, gathered, 1)

	metrics := gathered[0].GetMetric()
	require.Len(t, metrics, 2)

	countByMethod := make(map[string]uint64)
	for _, metric := range metrics {
		for _, lp := range metric.GetLabel() {
			if lp.GetName() == "method" {
				countByMethod[lp.GetValue()] = metric.GetHistogram().GetSampleCount()
			}
		}
	}
	assert.Equal(t, uint64(1), countByMethod["CmfTask.list"])
	assert.Equal(t, uint64(1), countByMethod["CmfProject.get"])
}


