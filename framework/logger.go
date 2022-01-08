package framework

import "github.com/labstack/echo/v4"

// making *echo.Echo private to prevent the object from being misused
type EchoLogger struct {
	e *echo.Echo
}

// Initializing Echo object
// Copies the address of the main echo object
// into the e in EchoLogger
func (ech *EchoLogger) Initialize(e *echo.Echo) {
	ech.e = e
}

func (ech EchoLogger) Error(message string) {
	ech.e.Logger.Error(message)
}

func (ech EchoLogger) Info(message string) {
	ech.e.Logger.Info(message)
}

func (ech EchoLogger) Warn(message string) {
	ech.e.Logger.Warn(message)
}

func (ech EchoLogger) Debug(message string) {
	ech.e.Logger.Debug(message)
}
