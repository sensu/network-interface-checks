//go:build linux
// +build linux

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	procNetDevInterfaceRE = regexp.MustCompile(`^(.+): *(.+)$`)
	procNetDevFieldSep    = regexp.MustCompile(` +`)

	metricTypeMap = map[string]*struct{ metricType, ingress, egress string }{
		"bytes":   {"bytes", "recv", "sent"},
		"drop":    {"drop", "in", "out"},
		"errs":    {"err", "in", "out"},
		"mtu":     {"mtu", "", ""},
		"packets": {"packets", "recv", "sent"},
	}
)

func getLocalInterfaceName() string {
	return "lo"
}

func GetNetStats(selector *selector) (NetStats, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	return parseNetStats(file, selector)
}

// parseNetStats queries the host and returns a map with the fillowing information:
// map[metricType] -> map[interface]value
func parseNetStats(r io.Reader, selector *selector) (NetStats, error) {
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first header
	scanner.Scan()
	parts := strings.Split(scanner.Text(), "|")
	if len(parts) != 3 { // interface + receive + transmit
		return nil, fmt.Errorf("invalid header line in net/dev: %s",
			scanner.Text())
	}

	receiveHeader := strings.Fields(parts[1])
	transmitHeader := strings.Fields(parts[2])
	receiveHeaderCount := len(receiveHeader)
	transmitHeaderCount := len(transmitHeader)
	headerCount := receiveHeaderCount + transmitHeaderCount

	statsByType := NetStats{}
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " ")
		if line == "" {
			continue
		}
		parts := procNetDevInterfaceRE.FindStringSubmatch(line)
		if len(parts) != 3 {
			return nil, fmt.Errorf("couldn't get interface metricType, invalid line in net/dev: %q", line)
		}

		dev := parts[1]
		if selector.Ignored(dev) {
			continue
		}

		values := procNetDevFieldSep.Split(strings.TrimLeft(parts[2], " "), -1)
		if len(values) != headerCount {
			return nil, fmt.Errorf("couldn't get values, invalid line in net/dev: %q", parts[2])
		}

		addStats := func(metricType string, ingress bool, value string) {
			metricMap := metricTypeMap[metricType]
			if metricMap == nil {
				return
			}
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return
			}

			postfix := metricMap.ingress
			if !ingress {
				postfix = metricMap.egress
			}

			metric := fmt.Sprintf("%s_%s", metricMap.metricType, postfix)
			statsForType, ok := statsByType[metric]
			if !ok {
				statsForType = map[string]float64{}
				statsByType[metric] = statsForType
			}

			statsForType[dev] = v
		}

		for i := 0; i < receiveHeaderCount; i++ {
			addStats(receiveHeader[i], true, values[i])
		}

		for i := 0; i < transmitHeaderCount; i++ {
			addStats(transmitHeader[i], false, values[i+receiveHeaderCount])
		}
	}
	return statsByType, scanner.Err()
}
