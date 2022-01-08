package main

import (
	"encoding/json"
	"lw-node-exporter/core"
	"lw-node-exporter/framework"
	"net/http"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	echologrus "github.com/spirosoik/echo-logrus"
)

func main() {

	e := echo.New()

	// Logrus Logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	//JSON Format // can set fluentD formats also
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Add the stack hook.
	// Custom setting caller level to panic, fatal and error
	logger.AddHook(logrus_stack.LogrusStackHook{
		CallerLevels: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel},
		StackLevels:  []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel},
	})

	// Writelogs function is a wrapper for calling Echo framework's logger
	// For centralized logging
	var writeLogs framework.EchoLogger
	writeLogs.Initialize(e)

	// Usage of logger middleware
	mw := echologrus.NewLoggerMiddleware(logger)
	e.Logger = mw
	e.Use(mw.Hook())

	// Need to add this function once in each package
	// This will ensure that the same logging function is reusable
	// acros all packages
	core.Initialize(writeLogs)

	e.GET("/", func(c echo.Context) error {

		// dockerRoot := "/Users/preetham/Documents/git/infra-utilities/node-exporter-lightweight"
		metricsOutput := core.GetCurrentInfo()
		jsonMetrics, err := json.Marshal(&metricsOutput)
		if err != nil {
			panic(err)
		}
		return c.String(http.StatusOK, string(jsonMetrics))

	})
	e.Logger.Fatal(e.Start(":1323"))
}
