package main

import (
	"bytes"
	"fmt"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	"log"
	"os"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Sum               bool
	IncludeInterfaces []string
	ExcludeInterfaces []string
}

const (
	cacheFile = "./.network-interface-check.json"
)

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "network-interface-checks",
			Short:    "Network Interface Checks",
			Keyspace: "sensu.io/plugins/network-interface-checks/config",
		},
		IncludeInterfaces: make([]string, 0),
		ExcludeInterfaces: make([]string, 0),
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "sum",
			Env:       "NETWORK_INTERFACE_CHECKS_SUM",
			Argument:  "sum",
			Shorthand: "s",
			Default:   false,
			Usage:     "Add additional measurement per metric w/ \"interface=all\" tag",
			Value:     &plugin.Sum,
		}, {
			Path:      "include-interfaces",
			Env:       "NETWORK_INTERFACE_CHECKS_INCLUDE_INTERFACES",
			Argument:  "include-interfaces",
			Shorthand: "i",
			Default:   []string{},
			Usage:     "Comma-delimited string of interface names to include",
			Value:     &plugin.IncludeInterfaces,
		}, {
			Path:      "exclude-interfaces",
			Env:       "NETWORK_INTERFACE_CHECKS_EXCLUDE_INTERFACES",
			Argument:  "exclude-interfaces",
			Shorthand: "x",
			Default:   []string{"lo"},
			Usage:     "Comma-delimited string of interface names to exclude",
			Value:     &plugin.ExcludeInterfaces,
		},
	}
)

func main() {
	useStdin := false
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Printf("Error check stdin: %v\n", err)
		panic(err)
	}
	//Check the Mode bitmask for Named Pipe to indicate stdin is connected
	if fi.Mode()&os.ModeNamedPipe != 0 {
		log.Println("using stdin")
		useStdin = true
	}

	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, useStdin)
	check.Execute()
}

func checkArgs(_ *v2.Event) (int, error) {
	fmt.Printf("%+v", plugin)

	return sensu.CheckStateOK, nil
}

func collectMetrics() ([]*dto.MetricFamily, error) {
	collector, err := NewCollector(plugin.IncludeInterfaces, plugin.ExcludeInterfaces, cacheFile)
	if err != nil {
		return nil, err
	}

	return collector.Collect(plugin.Sum, GetNetStats)
}

func executeCheck(_ *v2.Event) (int, error) {
	families, err := collectMetrics()
	if err != nil {
		return sensu.CheckStateCritical, err
	}

	var buf bytes.Buffer
	for _, family := range families {
		buf.Reset()
		encoder := expfmt.NewEncoder(&buf, expfmt.FmtText)
		err = encoder.Encode(family)
		if err != nil {
			return sensu.CheckStateCritical, err
		}

		fmt.Print(buf.String())
	}
	return sensu.CheckStateOK, nil
}
