package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log *Logger
)

type Logger struct {
	*zap.SugaredLogger
	logsPath string
	Name     string
}

// Config represents configuration options for logger initialization
type Config struct {
	Debug     bool   // Enable debug logging
	TimeZone  string // Set the time zone (GMT+0, GMT+3, etc.)
	LogToFile bool   // Enable logging to a file
	LogsDir   string // Set the directory for logs (default: current working directory)
}

// Init is a function to initialize logger with extended configuration
func Init(config Config) error {
	var l Logger
	l.Name = "main"

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Set log directory, default to current working directory
	if config.LogsDir == "" {
		l.logsPath = wd
	} else {
		l.logsPath = filepath.Join(wd, config.LogsDir)
	}

	// Ensure log directory exists
	err = os.MkdirAll(l.logsPath, os.ModePerm)
	if err != nil {
		return err
	}

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "timestamp",
		NameKey:        "logger",
		CallerKey:      "caller",
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	if config.TimeZone != "" {
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.In(time.FixedZone(config.TimeZone, 3*60*60)).Format("2006-01-02 15:04:05"))
		}
	}

	var level zapcore.Level
	if config.Debug {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	// Console encoder with colors
	consoleEncoderConfig := encoderConfig
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	// File encoder without colors
	fileEncoderConfig := encoderConfig
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	var cores []zapcore.Core

	// Add console output
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level)
	cores = append(cores, consoleCore)

	// Add file output if enabled
	if config.LogToFile {
		mainLogPath := filepath.Join(l.logsPath, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02 15:04")))
		fileWriter, errOpenFile := os.OpenFile(mainLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if errOpenFile != nil {
			return errOpenFile
		}

		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), level)
		cores = append(cores, fileCore)
	}

	// Create combined core
	combinedCore := zapcore.NewTee(cores...)
	log := zap.New(combinedCore, zap.AddCaller())

	l.SugaredLogger = log.Named(l.Name).Sugar()
	Log = &l

	return nil
}

// Named returns a new logger with the specified name ("bot", "database", etc.)
func Named(name string) (*Logger, error) {
	if Log == nil {
		return nil, fmt.Errorf("logger is not initialized")
	}
	return &Logger{
		SugaredLogger: Log.SugaredLogger.Named(name),
		logsPath:      Log.logsPath,
		Name:          name,
	}, nil
}

// customTimeEncoder formats time in GMT+0
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.In(time.FixedZone("GMT+0", 3*60*60)).Format("2006-01-02 15:04:05"))
}
