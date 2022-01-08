package core

import "lw-node-exporter/framework"

// Add this as part of any new packages in the future
// Logs function will be initialized via this package
// This allows for centralized logging
var writeLogs framework.EchoLogger

// Copies the value of struct
// framework.Echologger in main package
// into the writeLogs in current Package
func Initialize(tempLogFunction framework.EchoLogger) {
	writeLogs = tempLogFunction
}
