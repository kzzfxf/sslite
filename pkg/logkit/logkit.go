// Copyright 2023 kzzfxf
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logkit

import (
	"io"
	"log/slog"
)

type Level string

const (
	LevelDebug = Level("debug")
	LevelInfo  = Level("info")
	LevelWarn  = Level("warn")
	LevelError = Level("error")
)

var (
	logger *slog.Logger
)

// Init
func Init(w io.Writer, level Level) {
	l := slog.LevelError
	switch level {
	case LevelDebug:
		l = slog.LevelDebug
	case LevelInfo:
		l = slog.LevelInfo
	case LevelWarn:
		l = slog.LevelWarn
	case LevelError:
		l = slog.LevelError
	}
	logger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: l}))
}

type Attr [2]interface{}

// WithAttr
func WithAttr(field string, value interface{}) Attr {
	return Attr{field, value}
}

// attrs2args
func attrs2args(attrs ...Attr) (args []interface{}) {
	args = make([]interface{}, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr[0], attr[1])
	}
	return
}

// Debug
func Debug(msg string, attrs ...Attr) {
	if logger != nil {
		logger.Debug(msg, attrs2args(attrs...)...)
	}
}

// Info
func Info(msg string, attrs ...Attr) {
	if logger != nil {
		logger.Info(msg, attrs2args(attrs...)...)
	}
}

// Warn
func Warn(msg string, attrs ...Attr) {
	if logger != nil {
		logger.Warn(msg, attrs2args(attrs...)...)
	}
}

// Error
func Error(msg string, attrs ...Attr) {
	if logger != nil {
		logger.Error(msg, attrs2args(attrs...)...)
	}
}
