[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/network-interface-checks)
![Go Test](https://github.com/sensu/network-interface-checks/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/sensu/network-interface-checks/workflows/goreleaser/badge.svg)

# Sensu Network Interface Checks

## Table of Contents
- [Overview](#overview)
  - [Output Metrics](#output-metrics)
  - [Rate Metrics](#rate-metrics)
- [Usage examples](#usage-examples)
  - [Help output](#help-output)
  - [Environment variables](#environment-variables)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

The [Sensu Network Interface Checks][1] are Linux [Sensu Metrics Check][7] that provides baseline network metrics in prometheus format. 

### Output Metrics

| Name              | Type    | Description                         |
|-------------------|---------|-------------------------------------|
| bytes_sent        | counter | Bytes sent                          |
| bytes_sent_rate   | gauge   | Bytes sent per second               |
| bytes_recv        | counter | Bytes received                      |
| bytes_recv_rate   | gauge   | Bytes received per second           |
| packets_sent      | counter | Packets sent                        |
| packets_sent_rate | gauge   | Packets sent per second             |
| packets_recv      | counter | Packets received                    |
| packets_recv_rate | gauge   | Packets received per second         |
| err_out           | counter | Outbound errors                     |
| err_out_rate      | gauge   | Outbound errors per second          |
| err_in            | counter | Inbound errors                      |    
| err_in_rate       | gauge   | Inbound errors per second           |
| drop_out          | counter | Outbound packets dropped            |
| drop_out_rate     | gauge   | Outbound packets dropped per second |
| drop_in           | counter | Inbound packets dropped             |
| drop_in_rate      | gauge   | Inbound packets dropped per second  |

### Rate Metrics
In order to obtain rate metrics the `--state-file` argument must be used. The state file holds previous values and millisecond accurate timestamp, which are used to calculate metric rate using a simple time difference between current values and previously recorded values in the state file.  By default rate metrics are only calculated if the stored values in the selected state file are less than 60 seconds old. You can optionally set the maximum allowed time interval using `--max-rate-interval` if the 60 second default isn't suitable. 

If the state file does not exist or if the state is too stale, the rate metrics will not be produced. 
  
## Usage examples

### Help output

```
Network Interface Checks

Usage:
  network-interface-checks [flags]
  network-interface-checks [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -x, --exclude-interfaces strings   Comma-delimited string of interface names to exclude (default [lo])
  -h, --help                         help for network-interface-checks
  -i, --include-interfaces strings   Comma-delimited string of interface names to include
  -r, --max-rate-interval int        Maximum number of seconds since last measurement that triggers a rate calculation. 0 for no maximum. (default 60)
  -f, --state-file string            State file used for rate calculation. If empty no rate is calculated.
  -s, --sum                          Add additional measurement per metric w/ "interface=all" tag
      --sumologic-compat             Add Sumo Logic compatible metrics with w/ "host_net" family

Use "network-interface-checks [command] --help" for more information about a command.
```

### Environment variables
| Argument             | Environment Variable                        |
|----------------------|---------------------------------------------|
| --sum                | NETWORK_INTERFACE_CHECKS_SUM                |
| --include-interfaces | NETWORK_INTERFACE_CHECKS_INCLUDE_INTERFACES |
| --exclude-interfaces | NETWORK_INTERFACE_CHECKS_EXCLUDE_INTERFACES |
| --max-rate-interval  | NETWORK_INTERFACE_CHECKS_MAX_RATE_INTERVAL  |
| --state-file         | NETWORK_INTERFACE_CHECKS_STATE_FILE         |
| --sumologic-compat   | NETWORK_INTERFACE_CHECKS_SUMOLOGIC_COMPAT   |

## Configuration
### Asset registration

[Sensu Assets][11] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add sensu/network-interface-checks
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][12].

### Check definition

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: network-interface-checks
  namespace: default
spec:
  command: network-interface-checks
  subscriptions:
  - system
  runtime_assets:
  - sensu/network-interface-checks
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the network-interface-checks repository:

```
go build
```

## Additional notes

This plugin is only supported on Linux.

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/network-interface-checks
[2]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[3]: https://github.com/sensu/sensu-plugin-sdk
[4]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[5]: https://github.com/sensu/check-plugin-template/blob/master/.github/workflows/release.yml
[6]: https://github.com/sensu/check-plugin-template/actions
[7]: https://docs.sensu.io/sensu-go/latest/reference/checks/
[8]: https://github.com/sensu/check-plugin-template/blob/master/main.go
[9]: https://bonsai.sensu.io/
[10]: https://github.com/sensu/sensu-plugin-tool
[11]: https://docs.sensu.io/sensu-go/latest/plugins/assets/
[12]: https://bonsai.sensu.io/assets/sensu/network-interface-checks
