package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var (
	Log *Logger
	cfg Config
)

type Logger struct {
	*zap.SugaredLogger
	logsPath string
	prefix   string
}

// Config represents configuration options for logger initialization
type Config struct {
	Debug     bool
	TimeZone  string
	LogToFile bool
	LogsDir   string
}

// Init is a function to initialize logger with extended configuration
func Init(config Config) error {
	var l Logger
	l.prefix = "main"

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
		CallerKey:      "caller",
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	if config.TimeZone != "" {
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.In(time.FixedZone(config.TimeZone, 3*60*60)).Format("2006-0github.com/Badsnus/cu-clubs-bot-02 github.com/Badsnus/cu-clubs-bot5:04:05"))
		}
	}

	var level zapcore.Level
	if config.Debug {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	consoleEncoder := &prefixEncoder{
		Encoder: zapcore.NewConsoleEncoder(encoderConfig),
		pool:    buffer.NewPool(),
		prefix:  "[ MAIN ]",
	}
	var cores []zapcore.Core

	// Add console output
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level)
	cores = append(cores, consoleCore)

	// Add file output if enabled
	if config.LogToFile {
		mainLogPath := filepath.Join(l.logsPath, "main.log")
		fileWriter, errOpenFile := os.OpenFile(mainLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if errOpenFile != nil {
			return errOpenFile
		}

		fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(fileWriter), level)
		cores = append(cores, fileCore)
	}

	// Create combined core
	combinedCore := zapcore.NewTee(cores...)
	log := zap.New(combinedCore, zap.AddCaller())

	l.SugaredLogger = log.Sugar()
	Log = &l
	cfg = config

	return nil
}

func GetPrefixed(prefix string) (*Logger, error) {
	if Log == nil {
		return nil, fmt.Errorf("logger is not initialized")
	}

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "timestamp",
		CallerKey:      "caller",
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	if cfg.TimeZone != "" {
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.In(time.FixedZone(cfg.TimeZone, 3*60*60)).Format("2006-0github.com/Badsnus/cu-clubs-bot-02 github.com/Badsnus/cu-clubs-bot5:04:05"))
		}
	}

	var level zapcore.Level
	if cfg.Debug {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	consoleEncoder := &prefixEncoder{
		Encoder: zapcore.NewConsoleEncoder(encoderConfig),
		pool:    buffer.NewPool(),
		prefix:  fmt.Sprintf("[ %s ]", strings.ToUpper(prefix)),
	}
	var cores []zapcore.Core

	// Add console output
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level)
	cores = append(cores, consoleCore)

	// Add file output if enabled
	if cfg.LogToFile {
		mainLogPath := filepath.Join(Log.logsPath, fmt.Sprintf("%s.log", prefix))
		fileWriter, err := os.OpenFile(mainLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(fileWriter), level)
		cores = append(cores, fileCore)
	}

	// Create combined core
	combinedCore := zapcore.NewTee(cores...)
	log := zap.New(combinedCore, zap.AddCaller())

	return &Logger{
		SugaredLogger: log.Sugar(),
		logsPath:      Log.logsPath,
		prefix:        prefix,
	}, nil
}

func LogsPaths() ([]string, error) {
	files, err := os.ReadDir(Log.logsPath)
	if err != nil {
		return nil, err
	}

	var logs []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		logs = append(logs, filepath.Join(Log.logsPath, file.Name()))
	}

	return logs, nil
}

// customTimeEncoder formats time in GMT+0
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.In(time.FixedZone("GMT+0", 3*60*60)).Format("2006-0github.com/Badsnus/cu-clubs-bot-02 github.com/Badsnus/cu-clubs-bot5:04:05"))
}
