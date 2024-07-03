# mc-telemetry: A plentymarkets package for go telemetry

## Purpose: 

The objective of integrating this package is to enhance the existing logging system by implementing a more sophisticated framework, thereby improving the quality and depth of logged data.

The telemetry package is designed to facilitate logging into various third-party systems, including New Relic APM, New Relic ZeroLog, and Nope. Additionally, it enables straightforward console logging through the local driver.

This package includes interfaces for telemetry drivers, allowing the use of abstract interfaces rather than concrete implementations.

For more details about the available drivers, please refer to: [mc-telemetry-driver](..%2Fmc-telemetry-driver). 


## Getting Started:

To effectively utilize the Telemetry package, it is necessary to follow a series of steps, beginning with the installation of the appropriate package using the Go installer via the command line interface. This installation process ensures that all the required components of the Telemetry package are correctly set up.

Once the installation is complete, further configurations are essential to fully enable the functionality of the Telemetry package. Specifically, you will need to make adjustments in two key areas: the config.yaml file and the environment binder. These configurations are crucial as they define the operational parameters and environment settings necessary for Telemetry to function optimally.

The following structure needs to be updated:

```
my-app/
./
├── cmd/
│   ├── config.yaml
│   └── main.go
└── pkg/
    └── config.go
```

### Install:
Run command: 
`go get github.com/plentymarkets/mc-telemetry`

### Configuration:
Step 1. In main.go import the new packages: 
```go
package main

import (
    //...
	// Import the packages as follows.
	_ "github.com/plentymarkets/mc-telemetry-driver/pkg/teldrvr"
	"github.com/plentymarkets/mc-telemetry/pkg/telemetry" 
    //...
)

func main() { 
    //...
	// Configure the drivers 
	telemetry.SetDriver(strings.Split(cfg.GetString("telemetry.driver"), ",")...)
	telemetry.SetTraceDriver(cfg.GetString("telemetry.traceDriver"))
    // ...
}

```

Step 2. In config.yaml - Configure the packages as follows:

Note: This step is required only for development 
```yaml
# cmd/config.yaml
telemetry:
  driver: "local"       # The driver used for logging
  traceDriver: "local"  # The driver used for tracing between microservice
  app: "name-of-the-app"# Name of the microservice
  logLevel: "debug"     # Levels: error / info / debug
  newrelic:
    licenseKey: ""
```

Step 3. In the default configuration file, bind the environment variables 

```go
// GetConfig returns the configuration
func GetConfig(path string) (Config, error) {
	// telemetry
	viper.BindEnv("telemetry.driver", "TELEMETRY_DRIVER")
	viper.BindEnv("telemetry.traceDriver", "TELEMETRY_TRACE_DRIVER")
}

```

### Dynamic Configuration

For example, we can define the ExampleConfig as follows:

<table>
    <thead>
        <tr>
            <th>yaml</th>
            <th>env</th>
            <th>description</th>
            <th>default</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td><code>telemetry.driver</code></td>
            <td><code>TELEMETRY_DRIVER</code></td>
            <td>Thee trace driver used for logging:<br>
            Available drivers:
            <ul>
                <li>local - Run locally and log only to console</li>
                <li>newrelicAPM - Run locally and log only to console</li>
                <li>nrZerolog - Run locally and log only to console</li>
                <li>nopDriver - Run locally and log only to console</li>
            </ul>
            Note: One or more drivers can be used at once;
            </td>
            <td></td>
        </tr>
        <tr>
            <td><code>telemetry.traceeDriver</code></td>
            <td><code>TELEMETRY_TRACE_DRIVER</code></td>
            <td>Creates a traceID through which you can follow all the flow between microservices
            Available drivers:
            <ul>
                <li>local - Run locally and log only to console</li>
                <li>newrelicAPM - Run locally and log only to console</li>
                <li>nrZerolog - Run locally and log only to console</li>
                <li>nopDriver - Run locally and log only to console</li>
            </ul>
            Note: One or more drivers can be used at once;
            </td>
            <td></td>
        </tr>
        <tr>
            <td><code>telemetry.app</code></td>
            <td><code>TELEMETRY_APP</code></td>
            <td>The name of the microservice </td>
            <td></td>
        </tr>
        <tr>
            <td><code>telemetry.logLevel</code></td>
            <td><code>TELEMETRY_LOGLEVEL</code></td>
            <td> There are 3 types of log levels based on there priority
            <ul>
                <li><b>error</b> - Highest priority, logs only the errors</li>
                <li>info - Medium priority, logs both errors and info messages</li>
                <li>debug - Low priority, logs all levels plus extra info for debugging purposes</li>
            </ul>
            </td>
            <td>error</td>
        </tr>
        <tr>
            <td><code>telemetry.newrelic.licenseKey</code></td>
            <td><code>NEW_RELIC_LICENSE_KEY</code></td>
            <td>The licence key from newrelic</td>
            <td></td>
        </tr>
    </tbody>
</table>


## How to use telemetry:
```go
// Create the transaction object
transaction, err := telemetry.Start("Transaction Message")

if err != nil {
    log.Printf(err.Error())
}

defer transaction.Done() // Close transaction

// Start the Segment
segmentID := transaction.SegmentStart("Start handle user token")
defer transaction.SegmentEnd(segmentID) // Close segment

// Log message based on priority
transaction.Debug(segmentID, &msg)  // Requires segmentID string, msg *string
transaction.Info(segmentID, &msg)   // Requires segmentID string, msg *string
transaction.Error(segmentID, &err)  // Requires segmentID string, msg *error

```

## Dependencies: 
- go version >= 1.21
- [mc-telemetry-driver](https://github.com/plentymarkets/mc-telemetry-driverr)