package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDeviceSelector(t *testing.T) {
	tests := []struct {
		name      string
		includes  []string
		excludes  []string
		expectErr bool
	}{
		{
			name:      "empty includes and excludes",
			includes:  []string{},
			excludes:  []string{},
			expectErr: false,
		},
		{
			name:      "includes only",
			includes:  []string{"blah"},
			excludes:  []string{},
			expectErr: false,
		},
		{
			name:      "excludes only",
			includes:  []string{},
			excludes:  []string{"blah"},
			expectErr: false,
		},
		{
			name:      "both",
			includes:  []string{"doh"},
			excludes:  []string{"blah"},
			expectErr: true,
		},
		{
			name:      "both with local loop interface",
			includes:  []string{"doh"},
			excludes:  []string{"lo"},
			expectErr: false,
		},
		{
			name:      "both with local loop interface and other exclude",
			includes:  []string{"doh"},
			excludes:  []string{"lo", "blah"},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selector, err := NewDeviceSelector(test.includes, test.excludes)
			if test.expectErr {
				assert.Error(t, err)
				assert.Nil(t, selector)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, selector)
			}
		})
	}
}

func TestIgnored(t *testing.T) {
	tests := []struct {
		name     string
		includes []string
		excludes []string
		ifNames  []string
		ignored  []bool
	}{
		{
			name:     "empty includes and excludes",
			includes: []string{},
			excludes: []string{},
			ifNames:  []string{"en0", "en1", "en2", "en3"},
			ignored:  []bool{false, false, false, false},
		},
		{
			name:     "includes only",
			includes: []string{"en0", "en1"},
			excludes: []string{},
			ifNames:  []string{"en0", "en1", "en2", "en3"},
			ignored:  []bool{false, false, true, true},
		},
		{
			name:     "excludes only",
			includes: []string{},
			excludes: []string{"en0", "en1"},
			ifNames:  []string{"en0", "en1", "en2", "en3"},
			ignored:  []bool{true, true, false, false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selector, _ := NewDeviceSelector(test.includes, test.excludes)
			for ifIdx, ifName := range test.ifNames {
				assert.Equal(t, test.ignored[ifIdx], selector.Ignored(ifName), fmt.Sprintf("input: %s", ifName))
			}
		})
	}
}
