package telemetry

import (
	"io"
	"log"
	"strings"
)

// ErrorBytesSize is used for the default error size
const ErrorBytesSize = 1024

// Driver provides everything the application needs for telemetry
// Unfortunately go currently does nur support generics inside interface methods
type Driver interface {
	Start(string) Transaction
}

// Transaction ...
type Transaction interface {
	Logger
	AddAttribute(string, any)
	SegmentStart(string)
	SegmentEnd()
	Done()
}

// Logger ...
type Logger interface {
	Info(io.ReadCloser)
	Error(io.ReadCloser)
}

// registeredDriver holds all available driver
var registeredDriver map[string]Driver

// loadedDriver is a list of drivers to use for the application
var loadedDriver []string

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

// TransactionContainer ...
type TransactionContainer struct {
	transactions []Transaction
}

// StartAPM ...
func StartAPM(name string) TransactionContainer {
	transactionContainer := TransactionContainer{
		transactions: make([]Transaction, len(loadedDriver)),
	}

	for _, driverName := range loadedDriver {
		driver := getDriver(driverName)
		t := driver.Start(name)
		transactionContainer.transactions = append(transactionContainer.transactions, t)
	}

	return transactionContainer
}

// AddAttribute adds attributes to the registered driver transactions
func (tc *TransactionContainer) AddAttribute(name string, attribute any) {
	for _, transaction := range tc.transactions {
		transaction.AddAttribute(name, attribute)
	}
}

// SegmentStart starts a segment in the registered driver transactions
func (tc *TransactionContainer) SegmentStart(name string) {
	for _, transaction := range tc.transactions {
		transaction.SegmentStart(name)
	}
}

// SegmentEnd ends a segment in the registered driver transactions
func (tc *TransactionContainer) SegmentEnd() {
	for _, transaction := range tc.transactions {
		transaction.SegmentEnd()
	}
}

// Done ends the transactions for the registered driver
func (tc *TransactionContainer) Done() {
	for _, transaction := range tc.transactions {
		transaction.Done()
	}
}

// Info logs informations in the registered driver transactions
func (tc *TransactionContainer) Info(msg *string) {
	rc := io.NopCloser(strings.NewReader(*msg))

	for _, transaction := range tc.transactions {
		transaction.Info(rc)
	}
}

// Error logs errors in the registered driver transactions
func (tc *TransactionContainer) Error(err *error) {
	rc := io.NopCloser(strings.NewReader((*err).Error()))

	for _, transaction := range tc.transactions {
		transaction.Error(rc)
	}
}
