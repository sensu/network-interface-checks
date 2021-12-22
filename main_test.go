package main

import (
	"github.com/sensu/sensu-plugin-sdk/sensu"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckArgs(t *testing.T) {
	testCases := []struct {
		name              string
		includesIn        []string
		excludesIn        []string
		maxRateIntervalIn int64
		expectedStatus    int
		expectedError     bool
		expectedIncludes  []string
		expectedExcludes  []string
	}{
		{
			name:             "empty includes and excludes",
			includesIn:       []string{},
			excludesIn:       []string{},
			expectedStatus:   sensu.CheckStateOK,
			expectedError:    false,
			expectedIncludes: []string{},
			expectedExcludes: []string{},
		}, {
			name:             "includes only",
			includesIn:       []string{"eno1", "eno2"},
			excludesIn:       []string{},
			expectedStatus:   sensu.CheckStateOK,
			expectedError:    false,
			expectedIncludes: []string{"eno1", "eno2"},
			expectedExcludes: []string{},
		}, {
			name:             "includes only - trim spaces",
			includesIn:       []string{" eno1", "eno2 ", "  eno3 "},
			excludesIn:       []string{},
			expectedStatus:   sensu.CheckStateOK,
			expectedError:    false,
			expectedIncludes: []string{"eno1", "eno2", "eno3"},
			expectedExcludes: []string{},
		}, {
			name:             "excludes only",
			includesIn:       []string{},
			excludesIn:       []string{"eno1", "eno2"},
			expectedStatus:   sensu.CheckStateOK,
			expectedError:    false,
			expectedIncludes: []string{},
			expectedExcludes: []string{"eno1", "eno2"},
		}, {
			name:             "excludes only - trim spaces",
			includesIn:       []string{},
			excludesIn:       []string{"eno1 ", " eno2", " eno3 "},
			expectedStatus:   sensu.CheckStateOK,
			expectedError:    false,
			expectedIncludes: []string{},
			expectedExcludes: []string{"eno1", "eno2", "eno3"},
		}, {
			name:             "includes with local interface exclude",
			includesIn:       []string{"eno1", "eno2"},
			excludesIn:       []string{getLocalInterfaceName()},
			expectedStatus:   sensu.CheckStateOK,
			expectedError:    false,
			expectedIncludes: []string{"eno1", "eno2"},
			expectedExcludes: []string{},
		}, {
			name:             "includes with excludes",
			includesIn:       []string{"eno1", "eno2"},
			excludesIn:       []string{getLocalInterfaceName(), "docker0"},
			expectedStatus:   sensu.CheckStateCritical,
			expectedError:    true,
			expectedIncludes: nil,
			expectedExcludes: nil,
		}, {
			name:              "max rate interval 0",
			includesIn:        []string{},
			excludesIn:        []string{},
			maxRateIntervalIn: 0,
			expectedStatus:    sensu.CheckStateOK,
			expectedError:     false,
			expectedIncludes:  []string{},
			expectedExcludes:  []string{},
		}, {
			name:              "max rate interval positive",
			includesIn:        []string{},
			excludesIn:        []string{},
			maxRateIntervalIn: 10,
			expectedStatus:    sensu.CheckStateOK,
			expectedError:     false,
			expectedIncludes:  []string{},
			expectedExcludes:  []string{},
		}, {
			name:              "max rate interval negative",
			includesIn:        []string{},
			excludesIn:        []string{},
			maxRateIntervalIn: -10,
			expectedStatus:    sensu.CheckStateCritical,
			expectedError:     true,
			expectedIncludes:  []string{},
			expectedExcludes:  []string{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			plugin = Config{
				PluginConfig:           sensu.PluginConfig{},
				Sum:                    false,
				IncludeInterfaces:      testCase.includesIn,
				ExcludeInterfaces:      testCase.excludesIn,
				MaxRateIntervalSeconds: testCase.maxRateIntervalIn,
			}

			status, err := checkArgs(nil)
			assert.Equal(t, testCase.expectedStatus, status)
			if testCase.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedIncludes, plugin.IncludeInterfaces)
				assert.Equal(t, testCase.expectedExcludes, plugin.ExcludeInterfaces)
			}
		})
	}
}
