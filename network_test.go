package main

import (
	"github.com/google/uuid"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
	"time"
)

func GetNetStatsMock1(_ *selector) (NetStats, error) {
	return NetStats{
		"bytes_sent": map[string]float64{
			"eno1": 12345676,
			"eno2": 23435678,
		}, "err_in": map[string]float64{
			"eno1": 2,
			"eno2": 4,
		},
	}, nil
}

func GetNetStatsMock2(_ *selector) (NetStats, error) {
	return NetStats{
		"bytes_sent": map[string]float64{
			"eno1": 22345676,
			"eno2": 33435678,
		}, "err_in": map[string]float64{
			"eno1": 8,
			"eno2": 12,
		},
	}, nil
}

func TestMetricCollector_CollectWithSum(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), uuid.New().String()+".json")

	// Sum, no rates
	collector, err := NewCollector([]string{}, []string{}, true, tmpFile, 60)
	assert.NoError(t, err)
	families, err := collector.Collect(GetNetStatsMock1)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 2)
	familyMap := familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.Contains(t, familyMap, "err_in")
	for _, family := range families {
		assert.Len(t, family.Metric, 3)
		assert.True(t, hasSumMetric(family))
	}

	// Second run there will be sum and rates
	collector, err = NewCollector([]string{}, []string{}, true, tmpFile, 60)
	assert.NoError(t, err)
	families, err = collector.Collect(GetNetStatsMock2)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 4)
	familyMap = familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.Contains(t, familyMap, "bytes_sent_rate")
	assert.Contains(t, familyMap, "err_in")
	assert.Contains(t, familyMap, "err_in_rate")
	for _, family := range families {
		assert.Len(t, family.Metric, 3)
		assert.True(t, hasSumMetric(family))
	}
}

func TestMetricCollector_CollectNoSum(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), uuid.New().String()+".json")

	// Sum, no rates
	collector, err := NewCollector([]string{}, []string{}, false, tmpFile, 60)
	assert.NoError(t, err)
	families, err := collector.Collect(GetNetStatsMock1)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 2)
	familyMap := familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.Contains(t, familyMap, "err_in")
	for _, family := range families {
		assert.Len(t, family.Metric, 2)
		assert.False(t, hasSumMetric(family))
	}

	// Second run there will be sum and rates
	collector, err = NewCollector([]string{}, []string{}, false, tmpFile, 60)
	assert.NoError(t, err)
	families, err = collector.Collect(GetNetStatsMock2)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 4)
	familyMap = familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.Contains(t, familyMap, "bytes_sent_rate")
	assert.Contains(t, familyMap, "err_in")
	assert.Contains(t, familyMap, "err_in_rate")
	for _, family := range families {
		assert.Len(t, family.Metric, 2)
		assert.False(t, hasSumMetric(family))
	}
}

func TestCollect_NoFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), uuid.New().String()+".json")

	// No rates since it's first time writing to file
	collector, err := NewCollector([]string{}, []string{}, false, tmpFile, 60)
	assert.NoError(t, err)
	families, err := collector.Collect(GetNetStatsMock1)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 2)
	familyMap := familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.Contains(t, familyMap, "err_in")
	for _, family := range families {
		assert.Len(t, family.Metric, 2)
		assert.False(t, hasSumMetric(family))
	}

	// Second run with no file there will be no rates
	collector, err = NewCollector([]string{}, []string{}, false, "", 60)
	assert.NoError(t, err)
	families, err = collector.Collect(GetNetStatsMock2)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 2)
	familyMap = familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.NotContains(t, familyMap, "bytes_sent_rate")
	assert.Contains(t, familyMap, "err_in")
	assert.NotContains(t, familyMap, "err_in_rate")
}

func TestCollect_MaxRateInterval(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), uuid.New().String()+".json")

	// First run to generate a state file
	collector, err := NewCollector([]string{}, []string{}, false, tmpFile, 3)
	assert.NoError(t, err)
	families, err := collector.Collect(GetNetStatsMock1)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 2)
	familyMap := familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.Contains(t, familyMap, "err_in")
	for _, family := range families {
		assert.Len(t, family.Metric, 2)
		assert.False(t, hasSumMetric(family))
	}

	// Second run with a small delay, should produce rate metrics
	time.Sleep(time.Second * 1)
	collector, err = NewCollector([]string{}, []string{}, false, tmpFile, 3)
	assert.NoError(t, err)
	families, err = collector.Collect(GetNetStatsMock2)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 4)
	familyMap = familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.Contains(t, familyMap, "bytes_sent_rate")
	assert.Contains(t, familyMap, "err_in")
	assert.Contains(t, familyMap, "err_in_rate")

	// Third run with a long delay, no rates should be produced
	time.Sleep(time.Second * 3)
	collector, err = NewCollector([]string{}, []string{}, false, tmpFile, 1)
	assert.NoError(t, err)
	families, err = collector.Collect(GetNetStatsMock2)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 2)
	familyMap = familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.NotContains(t, familyMap, "bytes_sent_rate")
	assert.Contains(t, familyMap, "err_in")
	assert.NotContains(t, familyMap, "err_in_rate")

	// Fourth run with a long delay and 0 interval, should be produced
	time.Sleep(time.Second * 3)
	collector, err = NewCollector([]string{}, []string{}, false, tmpFile, 0)
	assert.NoError(t, err)
	families, err = collector.Collect(GetNetStatsMock2)
	assert.NoError(t, err)
	assert.NotNil(t, families)
	assert.Len(t, families, 4)
	familyMap = familiesByName(families)
	assert.Contains(t, familyMap, "bytes_sent")
	assert.Contains(t, familyMap, "bytes_sent_rate")
	assert.Contains(t, familyMap, "err_in")
	assert.Contains(t, familyMap, "err_in_rate")
}

func familiesByName(families []*dto.MetricFamily) map[string]*dto.MetricFamily {
	familyMap := map[string]*dto.MetricFamily{}
	for _, family := range families {
		familyMap[family.GetName()] = family
	}

	return familyMap
}

func hasSumMetric(family *dto.MetricFamily) bool {
	for _, metric := range family.GetMetric() {
		for _, label := range metric.GetLabel() {
			if label.GetName() == interfaceLabel && label.GetValue() == "all" {
				return true
			}
		}
	}
	return false
}
