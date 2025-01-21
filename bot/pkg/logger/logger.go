package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log     *types.Logger
	logHook types.LogHook
)

// Config represents configuration options for logger initialization
type Config struct {
	Debug     bool   // Enable debug logging
	TimeZone  string // Set the time zone (GMT+0, GMT+3, etc.)
	LogToFile bool   // Enable logging to a file
	LogsDir   string // Set the directory for logs (default: current working directory)
}

// SetLogHook sets a hook function that will be called for each log entry
func SetLogHook(hook types.LogHook) {
	logHook = hook
}

// customHook implements zapcore.Core to intercept log entries
type customHook struct {
	zapcore.Core
}

// Write intercepts the log entry and calls the hook function
func (h *customHook) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if logHook != nil {
		logHook(types.Log{
			Timestamp:  entry.Time,
			Caller:     entry.Caller.String(),
			LoggerName: entry.LoggerName,
			Level:      entry.Level.String(),
			Message:    entry.Message,
		}, fields)
	}
	return h.Core.Write(entry, fields)
}

// Init is a function to initialize logger with extended configuration
func Init(config Config) error {
	var l types.Logger
	l.Name = "main"

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Set log directory, default to current working directory
	if config.LogsDir == "" {
		l.LogsPath = wd
	} else {
		l.LogsPath = filepath.Join(wd, config.LogsDir)
	}

	// Ensure log directory exists
	err = os.MkdirAll(l.LogsPath, os.ModePerm)
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
		mainLogPath := filepath.Join(l.LogsPath, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02 15:04")))
		fileWriter, errOpenFile := os.OpenFile(mainLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if errOpenFile != nil {
			return errOpenFile
		}

		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), level)
		cores = append(cores, fileCore)
	}

	// Create combined core
	combinedCore := zapcore.NewTee(cores...)

	// Wrap the core with our custom hook
	hookedCore := &customHook{combinedCore}

	log := zap.New(hookedCore, zap.AddCaller())

	l.SugaredLogger = log.Named(l.Name).Sugar()
	Log = &l

	return nil
}

// Named returns a new logger with the specified name ("bot", "database", etc.)
func Named(name string) (*types.Logger, error) {
	if Log == nil {
		return nil, fmt.Errorf("logger is not initialized")
	}
	return &types.Logger{
		SugaredLogger: Log.SugaredLogger.Named(name),
		LogsPath:      Log.LogsPath,
		Name:          name,
	}, nil
}

// customTimeEncoder formats time in GMT+0
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.In(time.FixedZone("GMT+0", 3*60*60)).Format("2006-01-02 15:04:05"))
}
