package cache

import (
	"bytes"
	"fmt"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const (
	bufferError = "buffer-error"
	jsonCache   = `{"my_metric-label1=label1Value":{"value":1234.5678,"timestamp":1639776815123},"my_other_metric-label21=label21Value":{"value":456.789,"timestamp":1639666815456},"my_other_metric-label221=label22Value-label222=label222Value":{"value":987.21,"timestamp":1639666815789}}`
)

type ErrorReadWriter struct {
}

func (e ErrorReadWriter) Read(_ []byte) (n int, err error) {
	return 0, fmt.Errorf(bufferError)
}

func (e ErrorReadWriter) Write(_ []byte) (n int, err error) {
	return 0, fmt.Errorf(bufferError)
}

var (
	metricType = dto.MetricType_COUNTER

	family1Name           = "my_metric"
	family1Help           = "a beautiful metric"
	metric1Value          = 123.456
	metric1NewValue       = 1234.5678
	metric1TimestampMS    = int64(1639666815123)
	metric1NewTimestampMS = int64(1639776815123)
	metric1labelName      = "label1"
	metric1labelValue     = "label1Value"

	family2Name         = "my_other_metric"
	family2Help         = "another beautiful metric"
	metric21Value       = 456.789
	metric21TimestampMS = int64(1639666815456)
	metric21LabelName   = "label21"
	metric21LabelValue  = "label21Value"
	metric22Value       = 987.210
	metric22TimestampMS = int64(1639666815789)
	metric22LabelName1  = "label221"
	metric22LabelValue1 = "label22Value"
	metric22LabelName2  = "label222"
	metric22LabelValue2 = "label222Value"
)

func TestCounterMetricCache_AddGetMetric(t *testing.T) {
	family1 := &dto.MetricFamily{
		Name:   &family1Name,
		Help:   &family1Help,
		Type:   &metricType,
		Metric: []*dto.Metric{},
	}
	metric1 := &dto.Metric{
		Label: []*dto.LabelPair{{
			Name:  &metric1labelName,
			Value: &metric1labelValue,
		}},
		Counter: &dto.Counter{
			Value: &metric1Value,
		},
		TimestampMs: &metric1TimestampMS,
	}

	family2 := &dto.MetricFamily{
		Name:   &family2Name,
		Help:   &family2Help,
		Type:   &metricType,
		Metric: []*dto.Metric{},
	}
	metric21 := &dto.Metric{
		Label: []*dto.LabelPair{{
			Name:  &metric21LabelName,
			Value: &metric21LabelValue,
		}},
		Counter: &dto.Counter{
			Value: &metric21Value,
		},
		TimestampMs: &metric21TimestampMS,
	}
	metric22 := &dto.Metric{
		Label: []*dto.LabelPair{{
			Name:  &metric22LabelName1,
			Value: &metric22LabelValue1,
		}, {
			Name:  &metric22LabelName2,
			Value: &metric22LabelValue2,
		}},
		Counter: &dto.Counter{
			Value: &metric22Value,
		},
		TimestampMs: &metric22TimestampMS,
	}

	// AddMetric and GetMetric
	metricCache := New()
	metricCache.AddMetric(family1, metric1)
	metricCache.AddMetric(family2, metric21)
	metricCache.AddMetric(family2, metric22)

	found, value, timestampMS := metricCache.GetMetric(family1, metric1)
	assert.True(t, found)
	assert.Equal(t, metric1Value, value)
	assert.Equal(t, metric1TimestampMS, timestampMS)

	found, value, timestampMS = metricCache.GetMetric(family2, metric22)
	assert.True(t, found)
	assert.Equal(t, metric22Value, value)
	assert.Equal(t, metric22TimestampMS, timestampMS)

	// value update
	metric1.Counter.Value = &metric1NewValue
	metric1.TimestampMs = &metric1NewTimestampMS
	metricCache.AddMetric(family1, metric1)
	found, value, timestampMS = metricCache.GetMetric(family1, metric1)
	assert.True(t, found)
	assert.Equal(t, metric1NewValue, value)
	assert.Equal(t, metric1NewTimestampMS, timestampMS)

	// Write
	buf := new(bytes.Buffer)
	err := metricCache.Write(buf)
	assert.NoError(t, err)
	str := buf.String()
	assert.Equal(t, jsonCache, str)

	// Write error
	err = metricCache.Write(&ErrorReadWriter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), bufferError)

	// Read - make sure all 3 metrics are there
	metricCache = New()
	err = metricCache.Read(strings.NewReader(jsonCache))
	assert.NoError(t, err)
	found, value, timestampMS = metricCache.GetMetric(family1, metric1)
	assert.True(t, found)
	assert.Equal(t, metric1NewValue, value)
	assert.Equal(t, metric1NewTimestampMS, timestampMS)
	found, value, timestampMS = metricCache.GetMetric(family2, metric21)
	assert.True(t, found)
	assert.Equal(t, metric21Value, value)
	assert.Equal(t, metric21TimestampMS, timestampMS)
	found, value, timestampMS = metricCache.GetMetric(family2, metric22)
	assert.True(t, found)
	assert.Equal(t, metric22Value, value)
	assert.Equal(t, metric22TimestampMS, timestampMS)

	// Read error
	err = metricCache.Read(&ErrorReadWriter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), bufferError)
}
