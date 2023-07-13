package telemetry

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/google/uuid"
)

// ErrorBytesSize is used for the default error size
const ErrorBytesSize = 1024

// Default format for telemetry driver errors
const TelemetryDriverError = "Telemetry error in driver: "

// Driver provides everything the application needs for telemetry
// Unfortunately go currently does nur support generics inside interface methods
type Driver interface {
	Start(string) (Transaction, error)
}

// Transaction ...
type Transaction interface {
	Logger
	Tracer
	Allocator
	AddTransactionAttribute(string, any) error
	SegmentStart(string, string) error
	AddSegmentAttribute(string, string, any) error
	SegmentEnd(string) error
	Done() error
}

// Tracer ...
type Tracer interface {
	CreateTrace() (string, error)
	SetTrace(string) error
	Trace() (string, error)
}

// Logger ...
type Logger interface {
	Info(string, io.ReadCloser) error
	Error(string, io.ReadCloser) error
}

// Allocator ...
type Allocator interface {
	Erase()
}

// ErrorWrapper wrapps up multiple driver errors
type ErrorWrapper struct {
	errors []error
}

// registeredDriver holds all available driver
var registeredDriver map[string]Driver

// loadedDriver is a list of drivers to use for the application
var loadedDriver []string

// traceDriver is the driver used for the trace
var traceDriver string

// RegisterDriver adds the possibilty to add a driver to the driver map
func RegisterDriver(name string, driver Driver) {
	if registeredDriver == nil {
		registeredDriver = make(map[string]Driver)
	}

	registeredDriver[name] = driver
}

// getDriver returns the driver based on the provided name
func getDriver(name string) Driver {
	val, ok := registeredDriver[name]
	if !ok {
		log.Fatalf("provided telemetry driver is not registered. Driver name: %s", name)
	}

	return val
}

// SetDriver ...
func SetDriver(name ...string) {
	loadedDriver = name
}

// SetTraceDriver ...
func SetTraceDriver(name string) {
	traceDriver = name
}

// TransactionContainer ...
type TransactionContainer struct {
	transactions map[string]Transaction
}

// Start returns a transaction container with started transactions of all activated drivers.
func Start(name string) (TransactionContainer, error) {
	transactionContainer := TransactionContainer{
		transactions: make(map[string]Transaction, len(loadedDriver)),
	}

	for _, driverName := range loadedDriver {
		driver := getDriver(driverName)
		t, err := driver.Start(name)
		if err != nil {
			return transactionContainer, fmt.Errorf("%s%s - %w", TelemetryDriverError, driverName, err)
		}

		transactionContainer.transactions[driverName] = t
	}

	var trace string

	val, ok := transactionContainer.transactions[traceDriver]
	if !ok {
		return transactionContainer, fmt.Errorf("provided telemetry trace driver is not registered. Trace driver name: %s", traceDriver)
	}

	trace, err := val.CreateTrace()
	if err != nil {
		return transactionContainer, fmt.Errorf("%s%s\nFunction: CreateTrace\nError: %w", TelemetryDriverError, traceDriver, err)
	}

	err = transactionContainer.SetTrace(trace)
	if err != nil {
		return transactionContainer, err
	}

	return transactionContainer, nil
}

// AddTransactionAttribute adds attributes to the registered driver transactions
func (tc *TransactionContainer) AddTransactionAttribute(name string, attribute any) {
	for driverName, transaction := range tc.transactions {
		err := transaction.AddTransactionAttribute(name, attribute)
		if err != nil {
			log.Printf("%s%s\nFunction: AddTransactionAttribute\nError: %v", TelemetryDriverError, driverName, err)
		}
	}
}

// SegmentStart starts a segment in the registered driver transactions
func (tc *TransactionContainer) SegmentStart(name string) string {
	segmentID := uuid.NewString()

	for driverName, transaction := range tc.transactions {
		err := transaction.SegmentStart(segmentID, name)
		if err != nil {
			log.Printf("%s%s\nFunction: SegmentStart\nError: %v", TelemetryDriverError, driverName, err)
		}
	}

	return segmentID
}

// AddSegmentAttribute adds attributes to a segment for all driver
func (tc *TransactionContainer) AddSegmentAttribute(segmentID string, name string, attribute any) {
	for driverName, transaction := range tc.transactions {
		err := transaction.AddSegmentAttribute(segmentID, name, attribute)
		if err != nil {
			log.Printf("%s%s\nFunction: AddSegmentAttribute\nError: %v", TelemetryDriverError, driverName, err)
		}
	}
}

// SegmentEnd ends a segment in the registered driver transactions
func (tc *TransactionContainer) SegmentEnd(segmentID string) {
	for driverName, transaction := range tc.transactions {
		err := transaction.SegmentEnd(segmentID)
		if err != nil {
			log.Printf("%s%s\nFunction: SegmentEnd\nError: %v", TelemetryDriverError, driverName, err)
		}
	}
}

// SetTrace sets the trace for all transactions
func (tc *TransactionContainer) SetTrace(trace string) error {
	var ew ErrorWrapper

	for driverName, transaction := range tc.transactions {
		err := transaction.SetTrace(trace)
		if err != nil {
			ew.Add(fmt.Errorf("%s%s\nFunction: SetTrace\nError: %w", TelemetryDriverError, driverName, err))
		}
	}

	return ew.Error()
}

// Trace gets the trace of the transaction used for trace
func (tc *TransactionContainer) Trace() (string, error) {
	val, ok := tc.transactions[traceDriver]
	if !ok {
		return "", fmt.Errorf("provided telemetry trace driver is not registered. Trace driver name: %s", traceDriver)
	}

	trace, err := val.Trace()
	if err != nil {
		return "", fmt.Errorf("%s%s\nFunction: Trace\nError: %w", TelemetryDriverError, traceDriver, err)
	}

	return trace, nil
}

// Done ends the transactions for the registered driver
func (tc *TransactionContainer) Done() {
	for driverName, transaction := range tc.transactions {
		err := transaction.Done()
		if err != nil {
			log.Printf("%s%s\nFunction: Done\nError: %v", TelemetryDriverError, driverName, err)
		}
		transaction.Erase()
	}
}

// Info logs informations in the registered driver transactions
// If segmentID is empty, the info will be logged directly on the transaction
func (tc *TransactionContainer) Info(segmentID string, msg *string) {
	for driverName, transaction := range tc.transactions {
		rc := io.NopCloser(strings.NewReader(*msg))
		err := transaction.Info(segmentID, rc)
		if err != nil {
			log.Printf("%s%s\nFunction: Info\nError: %v", TelemetryDriverError, driverName, err)
		}
	}
}

// Error logs errors in the registered driver transactions
// If segmentID is empty, the error will be logged directly on the transaction
func (tc *TransactionContainer) Error(segmentID string, err *error) {
	for driverName, transaction := range tc.transactions {
		rc := io.NopCloser(strings.NewReader((*err).Error()))
		err := transaction.Error(segmentID, rc)
		if err != nil {
			log.Printf("%s%s\nFunction: Error\nError: %v", TelemetryDriverError, driverName, err)
		}
	}
}

// Error ...
func (ew *ErrorWrapper) Error() error {
	if len(ew.errors) == 0 {
		return nil
	}

	return errors.Join(ew.errors...)
}

// Add ...
func (ew *ErrorWrapper) Add(err error) {
	ew.errors = append(ew.errors, err)
}
