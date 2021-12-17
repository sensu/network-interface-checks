package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

const (
	netdev0 = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
`
	netdev1 = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
  eno1: 10858544415 11462995    123 1293763    66     77          88   1060490 501702477 5227309    76    98    0     0       0          0
`
	netdev3 = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 32921478  242896    12    34    55     66          77         88 32921477  242895    56    78    33     44       22          11
  eno1: 10858544415 11462995    123 1293763    66     77          88   1060490 501702477 5227309    76    98    0     0       0          0
tap-1e376645a40:  156002    1309    0    0    0     0          0         0  1280141    1311    0    0    0     0       0          0
`
)

var (
	result0 = NetStats{}
	result1 = NetStats{
		"bytes_sent":   {"eno1": 501702477},
		"bytes_recv":   {"eno1": 10858544415},
		"packets_sent": {"eno1": 5227309},
		"packets_recv": {"eno1": 11462995},
		"err_out":      {"eno1": 76},
		"err_in":       {"eno1": 123},
		"drop_out":     {"eno1": 98},
		"drop_in":      {"eno1": 1293763},
	}
	result3 = NetStats{
		"bytes_sent":   {"lo": 32921477, "eno1": 501702477, "tap-1e376645a40": 1280141},
		"bytes_recv":   {"lo": 32921478, "eno1": 10858544415, "tap-1e376645a40": 156002},
		"packets_sent": {"lo": 242895, "eno1": 5227309, "tap-1e376645a40": 1311},
		"packets_recv": {"lo": 242896, "eno1": 11462995, "tap-1e376645a40": 1309},
		"err_out":      {"lo": 56, "eno1": 76, "tap-1e376645a40": 0},
		"err_in":       {"lo": 12, "eno1": 123, "tap-1e376645a40": 0},
		"drop_out":     {"lo": 78, "eno1": 98, "tap-1e376645a40": 0},
		"drop_in":      {"lo": 34, "eno1": 1293763, "tap-1e376645a40": 0},
	}
)

func TestParseNetStats(t *testing.T) {
	baseSelector, _ := NewDeviceSelector([]string{}, []string{})
	includeSelector, _ := NewDeviceSelector([]string{"eno1"}, []string{})

	tests := []*struct {
		name           string
		input          string
		selector       *selector
		expectError    bool
		expectedResult NetStats
	}{
		{
			name:           "no interface",
			input:          netdev0,
			selector:       baseSelector,
			expectError:    false,
			expectedResult: result0,
		}, {
			name:           "one interface",
			input:          netdev1,
			selector:       baseSelector,
			expectError:    false,
			expectedResult: result1,
		}, {
			name:           "three interfaces",
			input:          netdev3,
			selector:       baseSelector,
			expectError:    false,
			expectedResult: result3,
		},
		{
			name:           "no interface, include selector",
			input:          netdev0,
			selector:       includeSelector,
			expectError:    false,
			expectedResult: result0,
		}, {
			name:           "one interface, include selector",
			input:          netdev1,
			selector:       includeSelector,
			expectError:    false,
			expectedResult: result1,
		}, {
			name:           "three interfaces, include selector",
			input:          netdev3,
			selector:       includeSelector,
			expectError:    false,
			expectedResult: result1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netStats, err := parseNetStats(strings.NewReader(test.input), test.selector)
			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, netStats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, netStats)
				assert.Equal(t, test.expectedResult, netStats)
			}
		})
	}
}
