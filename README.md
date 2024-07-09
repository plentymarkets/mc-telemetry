# mc-telemetry: A plentymarkets package for go telemetry

## Purpose

By integrating the mc-telemetry package we implement a more sophisticated framework. This enhances the existing logging system and improves the quality and depth of logged data.

The mc-telemetry package is designed to support logging into various third-party systems, including New Relic APM, New Relic ZeroLog, and Nope. Through the local driver the package also allows for straightforward console logging.

The mc-telemetry package includes interfaces for telemetry drivers, allowing users to use abstract interfaces rather than concrete implementations.

For more details about available drivers, please refer to: [mc-telemetry-driver](..%2Fmc-telemetry-driver). 


## Getting Started

To effectively use the mc-telemetry package, you first need to do the following: 

1. Install the appropriate package by using the *Go* installer via the command line interface.
   
   This ensures that all required components of the mc-telemetry package are correctly set up in your system.
   
2. Configure the *config.yaml* file and *environment binder* to unlock the full functionality of the mc-telemetry package.
    
	**_NOTE:_** These configurations are crucial to unlock the full functionality of the mc-telemetry package as they define necessary operational parameters and environment settings.

Update below structure as follows:

```
my-app/
./
├── cmd/
│   ├── config.yaml
│   └── main.go
└── pkg/
    └── config.go
```

### Installing the mc-telemetry package
Run: 

`go get github.com/plentymarkets/mc-telemetry`

### Importing the packages and configuring the drivers

**1.** In main.go import the new packages, like so:

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

**2.** To configure the packages in the *config.yaml* file, do the following:

**_Note:_** This is required only for development.

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

**3.** In the default configuration file, bind the environment variables:

```go
// GetConfig returns the configuration
func GetConfig(path string) (Config, error) {
	// telemetry
	viper.BindEnv("telemetry.driver", "TELEMETRY_DRIVER")
	viper.BindEnv("telemetry.traceDriver", "TELEMETRY_TRACE_DRIVER")
}

```

### Dynamic Configuration

For example, you can define the ExampleConfig as follows:

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
            Note: You can use one or more drivers at once
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
            Note: You can use one or more drivers at once
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
            <td> There are three types of log levels based on their priority
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


## How to use telemetry

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

## Dependencies

- go version >= 1.21
- [mc-telemetry-driver](https://github.com/plentymarkets/mc-telemetry-driverr)
