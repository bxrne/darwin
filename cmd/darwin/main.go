package main

import (
	"github.com/bxrne/logmgr"
)

func main() {
	defer logmgr.Shutdown()
	logmgr.AddSink(logmgr.DefaultConsoleSink) // Console output

	logmgr.Debug("This is a debug message")

}
