package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Sum                    bool
	SumoLogicCompat        bool
	IncludeInterfaces      []string
	ExcludeInterfaces      []string
	StateFile              string
	MaxRateIntervalSeconds int64
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "network-interface-checks",
			Short:    "Network Interface Checks",
			Keyspace: "sensu.io/plugins/network-interface-checks/config",
		},
		IncludeInterfaces:      make([]string, 0),
		ExcludeInterfaces:      make([]string, 0),
		StateFile:              "",
		MaxRateIntervalSeconds: 60,
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "sum",
			Env:       "NETWORK_INTERFACE_CHECKS_SUMOLOGIC_COMPAT",
			Argument:  "sumologic-compat",
			Shorthand: "",
			Default:   false,
			Usage:     "Add Sumo Logic compatible metrics with w/ \"host_net\" family",
			Value:     &plugin.SumoLogicCompat,
		}, {
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
			Default:   []string{getLocalInterfaceName()},
			Usage:     "Comma-delimited string of interface names to exclude",
			Value:     &plugin.ExcludeInterfaces,
		}, {
			Path:      "state-file",
			Env:       "NETWORK_INTERFACE_CHECKS_STATE_FILE",
			Argument:  "state-file",
			Shorthand: "f",
			Default:   "",
			Usage:     "State file used for rate calculation. If empty no rate is calculated.",
			Value:     &plugin.StateFile,
		}, {
			Path:      "max-rate-interval",
			Env:       "NETWORK_INTERFACE_CHECKS_MAX_RATE_INTERVAL",
			Argument:  "max-rate-interval",
			Shorthand: "r",
			Default:   int64(60),
			Usage:     "Maximum number of seconds since last measurement that triggers a rate calculation. 0 for no maximum.",
			Value:     &plugin.MaxRateIntervalSeconds,
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
	for i, include := range plugin.IncludeInterfaces {
		plugin.IncludeInterfaces[i] = strings.TrimSpace(include)
	}
	for i, exclude := range plugin.ExcludeInterfaces {
		plugin.ExcludeInterfaces[i] = strings.TrimSpace(exclude)
	}

	if len(plugin.IncludeInterfaces) > 0 && localInterfaceOnly(plugin.ExcludeInterfaces) {
		plugin.ExcludeInterfaces = []string{}
	}

	if len(plugin.IncludeInterfaces) > 0 && len(plugin.ExcludeInterfaces) > 0 && !localInterfaceOnly(plugin.ExcludeInterfaces) {
		return sensu.CheckStateCritical, fmt.Errorf("only one of --include-interfaces or --exclude-interfaces should be specified")
	}

	if plugin.MaxRateIntervalSeconds < 0 {
		return sensu.CheckStateCritical, fmt.Errorf("--max-rate-interval must be 0 or a positive value")
	}

	return sensu.CheckStateOK, nil
}

func collectMetrics() ([]*dto.MetricFamily, error) {
	collector, err := NewCollector(plugin.IncludeInterfaces, plugin.ExcludeInterfaces, plugin.Sum, plugin.SumoLogicCompat, plugin.StateFile,
		plugin.MaxRateIntervalSeconds)
	if err != nil {
		return nil, err
	}

	return collector.Collect(GetNetStats)
}

func executeCheck(_ *v2.Event) (int, error) {
	err := generateMetrics()
	if err != nil {
		fmt.Printf("Error executing %s: %v\n", plugin.Name, err)
		return sensu.CheckStateCritical, nil
	}
	return sensu.CheckStateOK, nil
}

func generateMetrics() error {
	families, err := collectMetrics()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	for _, family := range families {
		buf.Reset()
		encoder := expfmt.NewEncoder(&buf, expfmt.FmtText)
		err = encoder.Encode(family)
		if err != nil {
			return err
		}

		fmt.Print(buf.String())
	}

	return nil
}

func localInterfaceOnly(ifs []string) bool {
	return len(ifs) == 1 && ifs[0] == getLocalInterfaceName()
}
