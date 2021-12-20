package main

import (
	"fmt"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/log"
	"github.com/sensu/network-interface-checks/cache"
	"os"
	"time"
)

var (
	metricHelp = map[string]string{
		"bytes_sent":        "bytes sent",
		"bytes_sent_rate":   "bytes sent per second",
		"bytes_recv":        "bytes received",
		"bytes_recv_rate":   "bytes received per second",
		"packets_sent":      "packets sent",
		"packets_sent_rate": "packets sent per second",
		"packets_recv":      "packets received",
		"packets_recv_rate": "packets received per second",
		"err_out":           "outbound errors",
		"err_out_rate":      "outbound errors per second",
		"err_in":            "inbound errors",
		"err_in_rate":       "inbound errors per second",
		"drop_out":          "outbound packets dropped",
		"drop_out_rate":     "outbound packets dropped per second",
		"drop_in":           "incoming packets dropped",
		"drop_in_rate":      "incoming packets dropped per second",
		"mtu":               "interface MTU configuration",
	}
	interfaceLabel = "interface"
)

type MetricCollector struct {
	selector  *selector
	cacheFile string
}

// NetStats is the following: map[metric-name]map[interface-name]value
type NetStats map[string]map[string]float64

func NewCollector(includes, excludes []string, cacheFile string) (*MetricCollector, error) {
	selector, err := NewDeviceSelector(includes, excludes)
	if err != nil {
		return nil, err
	}

	return &MetricCollector{selector, cacheFile}, nil
}

func (c *MetricCollector) Collect(sum bool, netStatsGetter func(*selector) (NetStats, error)) ([]*dto.MetricFamily, error) {
	stats, err := netStatsGetter(c.selector)
	if err != nil {
		return nil, fmt.Errorf("couldn't get netstats: %w", err)
	}

	metricCache, err := cache.NewFromFile(c.cacheFile)
	if err != nil {
		log.Warnf("error opening cache file %s, continuing without rate metrics", cacheFile)
	}

	families := generatePromMetrics(stats, metricCache, sum)

	outFile, err := os.OpenFile(c.cacheFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Warnf("error writing cache file %s: %v", cacheFile, err)
		return families, err
	}
	defer func() { _ = outFile.Close() }()
	err = metricCache.Write(outFile)
	if err != nil {
		return families, err
	}

	return families, nil
}

func generatePromMetrics(stats NetStats, metricCache *cache.CounterMetricCache, sum bool) []*dto.MetricFamily {
	families := make([]*dto.MetricFamily, 0)
	nowMS := time.Now().UnixMilli()

	for metricType, typeStats := range stats {
		help := metricHelp[metricType]
		if help == "" {
			help = fmt.Sprintf("Network interface statistic %s.", metricType)
		}
		family := newMetricFamily(metricType, help, dto.MetricType_COUNTER)
		families = append(families, family)

		rateMetricType := metricType + "_rate"
		rateHelp := metricHelp[rateMetricType]
		if rateHelp == "" {
			rateHelp = fmt.Sprintf("Network interface %s per second.", metricType)
		}
		rateFamily := newMetricFamily(rateMetricType, rateHelp, dto.MetricType_GAUGE)

		var total float64 = 0
		var rateTotal float64 = 0
		hasRate := false

		for netIF, ifValue := range typeStats {
			metric := newCounterMetric(family, netIF, ifValue, nowMS)
			found, prevValue, prevTimestampMS := metricCache.GetMetric(family, metric)
			metricCache.AddMetric(family, metric)
			total += ifValue

			if found {
				rate := (ifValue - prevValue) / float64((nowMS-prevTimestampMS)/1000)
				newGaugeMetric(rateFamily, netIF, rate, nowMS)
				rateTotal += rate
				hasRate = true
			}
		}

		if hasRate {
			families = append(families, rateFamily)
		}

		if sum {
			newCounterMetric(family, "all", total, nowMS)
			if hasRate {
				newGaugeMetric(rateFamily, "all", rateTotal, nowMS)
			}
		}
	}

	return families
}

func newMetricFamily(name, help string, metricType dto.MetricType) *dto.MetricFamily {
	return &dto.MetricFamily{
		Name:   &name,
		Help:   &help,
		Type:   &metricType,
		Metric: []*dto.Metric{},
	}
}

func newCounterMetric(family *dto.MetricFamily, ifName string, value float64, timestampMS int64) *dto.Metric {
	metric := &dto.Metric{
		Label: []*dto.LabelPair{{Name: &interfaceLabel, Value: &ifName}},
		Counter: &dto.Counter{
			Value: &value,
		},
		TimestampMs: &timestampMS,
	}
	family.Metric = append(family.Metric, metric)

	return metric
}

func newGaugeMetric(family *dto.MetricFamily, ifName string, value float64, timestampMS int64) *dto.Metric {
	metric := &dto.Metric{
		Label: []*dto.LabelPair{{Name: &interfaceLabel, Value: &ifName}},
		Gauge: &dto.Gauge{
			Value: &value,
		},
		TimestampMs: &timestampMS,
	}
	family.Metric = append(family.Metric, metric)

	return metric
}