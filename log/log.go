// Copyright (c) 2021 mobus sunsc0220@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"fmt"
	stdlog "log"
	"os"
)

// Logger is a generic logging interface
type Logger interface {
	Log(v ...interface{})
	Logf(format string, v ...interface{})
}

// Level is a log level
type Level int

const (
	// LevelFatal fatal level
	LevelFatal Level = iota + 1
	// LevelError error level
	LevelError
	// LevelInfo info level
	LevelInfo
	// LevelDebug debug level
	LevelDebug
)

func (l Level) String() string {
	switch l {
	default:
		return ""
	case LevelDebug:
		return "[DBG]"
	case LevelError:
		return "[ERR]"
	case LevelInfo:
		return "[INF]"
	case LevelFatal:
		return "[FTL]"
	}
}

var (
	// the local logger
	logger Logger = &defaultLogLogger{}

	// default log level is info
	level = LevelInfo

	prefix = "[Slog]"
)

type defaultLogLogger struct{}

func (t *defaultLogLogger) Log(v ...interface{}) {
	stdlog.Print(v...)
}

func (t *defaultLogLogger) Logf(format string, v ...interface{}) {
	stdlog.Printf(format, v...)
}

func init() {
	switch os.Getenv("SLOG_LOG_LEVEL") {
	case "debug":
		level = LevelDebug
	case "info":
		level = LevelInfo
	case "error":
		level = LevelError
	case "fatal":
		level = LevelFatal
	}
}

func SetLogPrefix(logName string) {
	prefix = logName
}

// Log makes use of Logger
func Log(v ...interface{}) {
	if len(prefix) > 0 {
		logger.Log(append([]interface{}{prefix, " "}, v...)...)
		return
	}
	logger.Log(v...)
}

// Logf makes use of Logger
func Logf(format string, v ...interface{}) {
	if len(prefix) > 0 {
		format = prefix + " " + format
	}
	logger.Logf(format, v...)
}

// WithLevel logs with the level specified
func WithLevel(l Level, v ...interface{}) {
	if l > level {
		return
	}
	Log(v...)
}

// WithLevelf logs with the level specified
func WithLevelf(l Level, format string, v ...interface{}) {
	if l > level {
		return
	}
	Logf(format, v...)
}

// Debug provides debug level logging
func Debug(v ...interface{}) {
	WithLevel(LevelDebug, v...)
}

// Debugf provides debug level logging
func Debugf(format string, v ...interface{}) {
	WithLevelf(LevelDebug, format, v...)
}

// Info provides info level logging
func Info(v ...interface{}) {
	WithLevel(LevelInfo, v...)
}

// Infof provides info level logging
func Infof(format string, v ...interface{}) {
	WithLevelf(LevelInfo, format, v...)
}

// Error provides warn level logging
func Error(v ...interface{}) {
	WithLevel(LevelError, v...)
}

// Errorf provides warn level logging
func Errorf(format string, v ...interface{}) {
	WithLevelf(LevelError, format, v...)
}

// Fatal logs with Log and then exits with os.Exit(1)
func Fatal(v ...interface{}) {
	WithLevel(LevelFatal, v...)
	os.Exit(1)
}

// Fatalf logs with Logf and then exits with os.Exit(1)
func Fatalf(format string, v ...interface{}) {
	WithLevelf(LevelFatal, format, v...)
	os.Exit(1)
}

// SetLogger sets the local logger
func SetLogger(l Logger) {
	logger = l
}

// GetLogger returns the local logger
func GetLogger() Logger {
	return logger
}

// SetLevel sets the log level
func SetLevel(l Level) {
	level = l
}

// GetLevel returns the current level
func GetLevel() Level {
	return level
}

// SetPrefix sets a prefix for the logger
func SetPrefix(p string) {
	prefix = p
}

// Name sets service name
func Name(name string) {
	prefix = fmt.Sprintf("[%s]", name)
}
