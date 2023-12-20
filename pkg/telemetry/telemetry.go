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

// ErrorProcessID ...
type ErrorProcessID struct {
	err error
}

// Error returns the wrapped process id error
func (ep ErrorProcessID) Error() string {
	return fmt.Sprintf("ProcessID error. Err: %w", ep.err)
}

// Driver provides everything the application needs for telemetry
// Unfortunately go currently does nur support generics inside interface methods
type Driver interface {
	InitializeTransaction(string) (Transaction, error)
}

// Transaction ...
type Transaction interface {
	Logger
	Tracer
	Allocator
	Processor
	Start(string)
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

// Processor ...
type Processor interface {
	CreateProcessID() (string, error)
	SetProcessID(string) error
	ProcessID() (string, error)
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
		t, err := driver.InitializeTransaction(name)
		if err != nil {
			return transactionContainer, fmt.Errorf("%s%s - %w", TelemetryDriverError, driverName, err)
		}

		transactionContainer.transactions[driverName] = t
	}

	processID, err := transactionContainer.CreateProcessID()
	if err != nil {
		return transactionContainer, ErrorProcessID{
			err: err,
		}
	}

	err = transactionContainer.SetProcessID(processID)
	if err != nil {
		return transactionContainer, ErrorProcessID{
			err: err,
		}
	}

	for _, transaction := range transactionContainer.transactions {
		transaction.Start(name)
	}

	return transactionContainer, nil
}

// CreateProcessID creates the process id for all drivers depending on the trace driver
func (tc *TransactionContainer) CreateProcessID() (string, error) {
	var processID string
	val, ok := tc.transactions[traceDriver]
	if !ok {
		return processID, fmt.Errorf("provided telemetry trace driver is not registered. Trace driver name: %s", traceDriver)
	}

	processID, err := val.CreateProcessID()
	if err != nil {
		return processID, fmt.Errorf("%s%s Function: CreateProcessID | Error: %w", TelemetryDriverError, traceDriver, err)
	}

	return processID, nil
}

// StartTracing creates and sets the trace for all drivers depending on the trace driver
func (tc *TransactionContainer) StartTracing() (string, error) {
	var trace string

	val, ok := tc.transactions[traceDriver]
	if !ok {
		return trace, fmt.Errorf("provided telemetry trace driver is not registered. Trace driver name: %s", traceDriver)
	}

	trace, err := val.CreateTrace()
	if err != nil {
		return trace, fmt.Errorf("%s%s Function: CreateTrace | Error: %w", TelemetryDriverError, traceDriver, err)
	}

	err = tc.SetTrace(trace)
	if err != nil {
		return trace, err
	}

	return trace, nil
}

// AddTransactionAttribute adds attributes to the registered driver transactions
func (tc *TransactionContainer) AddTransactionAttribute(name string, attribute any) {
	for driverName, transaction := range tc.transactions {
		err := transaction.AddTransactionAttribute(name, attribute)
		if err != nil {
			log.Printf("%s%s Function: AddTransactionAttribute | Error: %v", TelemetryDriverError, driverName, err)
		}
	}
}

// SegmentStart starts a segment in the registered driver transactions
func (tc *TransactionContainer) SegmentStart(name string) string {
	segmentID := uuid.NewString()

	for driverName, transaction := range tc.transactions {
		err := transaction.SegmentStart(segmentID, name)
		if err != nil {
			log.Printf("%s%s Function: SegmentStart | Error: %v", TelemetryDriverError, driverName, err)
		}
	}

	return segmentID
}

// AddSegmentAttribute adds attributes to a segment for all driver
func (tc *TransactionContainer) AddSegmentAttribute(segmentID string, name string, attribute any) {
	for driverName, transaction := range tc.transactions {
		err := transaction.AddSegmentAttribute(segmentID, name, attribute)
		if err != nil {
			log.Printf("%s%s Function: AddSegmentAttribute | Error: %v", TelemetryDriverError, driverName, err)
		}
	}
}

// SegmentEnd ends a segment in the registered driver transactions
func (tc *TransactionContainer) SegmentEnd(segmentID string) {
	for driverName, transaction := range tc.transactions {
		err := transaction.SegmentEnd(segmentID)
		if err != nil {
			log.Printf("%s%s Function: SegmentEnd | Error: %v", TelemetryDriverError, driverName, err)
		}
	}
}

// SetProcessID sets the trace for all transactions
func (tc *TransactionContainer) SetProcessID(processID string) error {
	var ew ErrorWrapper

	for driverName, transaction := range tc.transactions {
		err := transaction.SetProcessID(processID)
		if err != nil {
			ew.Add(fmt.Errorf("%s%s Function: SetProcessID | Error: %w", TelemetryDriverError, driverName, err))
		}
	}

	return ew.Error()
}

// SetTrace sets the trace for all transactions
func (tc *TransactionContainer) SetTrace(trace string) error {
	var ew ErrorWrapper

	for driverName, transaction := range tc.transactions {
		err := transaction.SetTrace(trace)
		if err != nil {
			ew.Add(fmt.Errorf("%s%s Function: SetTrace | Error: %w", TelemetryDriverError, driverName, err))
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
		return "", fmt.Errorf("%s%s Function: Trace | Error: %w", TelemetryDriverError, traceDriver, err)
	}

	return trace, nil
}

// Done ends the transactions for the registered driver
func (tc *TransactionContainer) Done() {
	for driverName, transaction := range tc.transactions {
		err := transaction.Done()
		if err != nil {
			log.Printf("%s%s Function: Done | Error: %v", TelemetryDriverError, driverName, err)
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
			log.Printf("%s%s | Function: Info | Error: %v", TelemetryDriverError, driverName, err)
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
			log.Printf("%s%s Function: Error | Error: %v", TelemetryDriverError, driverName, err)
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
