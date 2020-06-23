package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger initializes then returns a zap.Logger.
func NewZapLogger(path string) *zap.Logger {
	dir := filepath.Dir(path)
	if exist, _ := exists(dir); !exist {
		if err := os.MkdirAll(dir, 0775); err != nil {
			panic(fmt.Sprintf("Fail to create log directory: %v", err))
		}
	}

	_, close, err := zap.Open(path)
	if err != nil {
		close()
		panic(fmt.Sprintf("Fail to create/open log file: %v", err))
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	atom := zap.NewAtomicLevelAt(zap.DebugLevel)

	config := zap.Config{
		Level:            atom,
		Development:      true,
		//DisableCaller: false,
		//DisableStacktrace: false,
		//Sampling: nil,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{path},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]interface{}{"serviceName": "ngapp"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("Fail to initialize logger: %v", err))
	}

	defer logger.Sync()
	logger.Info("Logger initialized.")

	return logger
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

