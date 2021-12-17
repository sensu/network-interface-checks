package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	dto "github.com/prometheus/client_model/go"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type CounterMetric struct {
	Value       float64 `json:"value"`
	TimestampMS int64   `json:"timestamp"`
}

type CounterMetricCache struct {
	metrics map[string]*CounterMetric
}

func New() *CounterMetricCache {
	return &CounterMetricCache{
		metrics: make(map[string]*CounterMetric),
	}
}

// NewFromFile creates a new cache and populates it with the content of the specified JSON file.
func NewFromFile(filename string) (*CounterMetricCache, error) {
	metricCache := New()

	info, err := os.Stat(filename)
	if (err != nil && !errors.Is(err, os.ErrNotExist)) || (info != nil && info.IsDir()) {
		return metricCache, fmt.Errorf("unable to use cache file %s", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		// ignore error, it just means the file doesn't exist
		return metricCache, nil
	}
	defer func() { _ = file.Close() }()

	err = metricCache.Read(file)
	if err != nil {
		return nil, fmt.Errorf("error reading cache file %s: %v", filename, err)
	}

	return metricCache, nil
}

func (s *CounterMetricCache) AddMetric(family *dto.MetricFamily, metric *dto.Metric) {
	key := getMetricKey(family, metric)
	s.metrics[key] = &CounterMetric{
		Value:       metric.GetCounter().GetValue(),
		TimestampMS: metric.GetTimestampMs(),
	}
}

func (s *CounterMetricCache) GetMetric(family *dto.MetricFamily, metric *dto.Metric) (bool, float64, int64) {
	key := getMetricKey(family, metric)
	cachedMetric := s.metrics[key]
	if cachedMetric == nil {
		return false, 0, 0
	}
	return true, cachedMetric.Value, cachedMetric.TimestampMS
}

func (s *CounterMetricCache) Write(writer io.Writer) error {
	content, err := json.Marshal(s.metrics)
	if err != nil {
		return fmt.Errorf("error creating json document: %v", err)
	}

	_, err = writer.Write(content)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func (s *CounterMetricCache) Read(reader io.Reader) error {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading cache content: %v", err)
	}
	err = json.Unmarshal(content, &s.metrics)
	if err != nil {
		return fmt.Errorf("error unmarshalling json cache content: %v", err)
	}
	return nil
}

func getMetricKey(family *dto.MetricFamily, metric *dto.Metric) string {
	var key strings.Builder
	key.WriteString(family.GetName())
	for _, label := range metric.GetLabel() {
		key.WriteString("-")
		key.WriteString(label.GetName())
		key.WriteString("=")
		key.WriteString(label.GetValue())
	}
	return key.String()
}
