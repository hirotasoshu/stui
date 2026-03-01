package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger(debug bool) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)
	logPath := filepath.Join(exeDir, "stui.log")

	level := zapcore.ErrorLevel
	if debug {
		level = zapcore.DebugLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{logPath},
		ErrorOutputPaths: []string{logPath},
	}

	Logger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
