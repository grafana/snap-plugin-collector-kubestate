/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Fields map[string]interface{}

var logFile *os.File
var err error

func init() {

	// Log as default ASCII formatter
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// try to open file for logging
	fname := strings.Join([]string{time.Now().Format("2006-01-02"), filepath.Base(os.Args[0])}, "_") + ".log"
	if logFile == nil {
		logFile, err = os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Println("Logging to stderr")
		}
	}

	// Output to file or to stderr
	if err != nil {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(logFile)
	}

	// log PLUGIN_LOG_LEVEL severity or above.
	switch strings.ToLower(os.Getenv("PLUGIN_LOG_LEVEL")) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.ErrorLevel)
	}
}

func LogInfo(message string, args ...interface{}) {
	setEntry(args...).Info(message)
}

func LogWarn(message string, args ...interface{}) {
	setEntry(args...).Warn(message)
}

func LogDebug(message string, args ...interface{}) {
	setEntry(args...).Debug(message)
}

func LogError(message string, args ...interface{}) {
	setEntry(args...).Error(message)
}

func LogFatal(message string, args ...interface{}) {
	setEntry(args...).Fatal(message)
}

func LogPanic(message string, args ...interface{}) {
	setEntry(args...).Panic(message)
}

func Log(fields map[string]interface{}) *log.Entry {
	fields["_func"] = getFunctionName(2)
	return log.WithFields(fields)
}

func setEntry(args ...interface{}) *log.Entry {
	fields := log.Fields{"_func": getFunctionName(3)}

	switch len(args) {
	case 0:
	case 1:
		fields["val"] = args[0]
	case 2:
		key := args[0].(string)
		fields[key] = args[1]
	default:
		fields["vals"] = args
	}

	return log.WithFields(fields)
}

func getFunctionName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)

	if !ok {
		return "!<not_accessible>"
	}

	return filepath.Base(runtime.FuncForPC(pc).Name())
}
