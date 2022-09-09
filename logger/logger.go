package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	envLogLevel  = "LOG_LEVEL"
	envLogOutput = "LOG_OUTPUT"
)

var (
	log logger
)

type bookstoreLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})

	// Elastic Search
	LogRoundTrip(*http.Request, *http.Response, error, time.Time, time.Duration) error
	RequestBodyEnabled() bool
	ResponseBodyEnabled() bool
}

type logger struct {
	log *zap.Logger
}

func init() {
	logConfig := zap.Config{
		OutputPaths: []string{getOutput()},
		Level:       zap.NewAtomicLevelAt(getLevel()),
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			LevelKey:     "level",
			TimeKey:      "time",
			MessageKey:   "msg",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	var err error
	if log.log, err = logConfig.Build(); err != nil {
		panic(err)
	}
}

func getLevel() zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(envLogLevel))) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

func getOutput() string {
	output := strings.TrimSpace(os.Getenv(envLogOutput))
	if output == "" {
		return "stdout"
	}
	return output
}

func GetLogger() bookstoreLogger {
	return log
}

func (l logger) Printf(format string, v ...interface{}) {
	if len(v) == 0 {
		Info(format, []string{})
	} else {
		Info(fmt.Sprintf(format, v...), []string{})
	}
}

func (l logger) Print(v ...interface{}) {
	Info(fmt.Sprintf("%v", v), []string{})
}

func (l logger) LogRoundTrip(request *http.Request, response *http.Response, err error, time time.Time, duration time.Duration) error {
	Info(
		"elastic",
		[]string{"request", "response", "error", "time", "duration"},
		request,
		response,
		err,
		time,
		duration,
	)
	return nil
}

func (l logger) RequestBodyEnabled() bool {
	return true
}

func (l logger) ResponseBodyEnabled() bool {
	return true
}

func Info(msg string, tagNames []string, tags ...interface{}) {
	log.log.Info(msg, mapFields(tagNames, tags)...)
	log.log.Sync()
}

func Error(msg string, err error, tagNames []string, tags ...interface{}) {
	tags = append(tags, zap.NamedError("error", err))
	log.log.Error(msg, mapFields(tagNames, tags)...)
	log.log.Sync()
}

func mapFields(tagNames []string, tags []interface{}) []zap.Field {
	fields := make(
		[]zap.Field,
		int64(math.Min(
			float64(len(tagNames)),
			float64(len(tags)),
		)),
	)
	for i := range fields {
		fields[i] = zap.Any(tagNames[i], tags[i])
	}
	return fields
}
