package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger interface defines logging methods
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// LogrusLogger is a wrapper around logrus.Logger
type LogrusLogger struct {
	logger *logrus.Entry
}

// NewLogger creates a new logger instance
func NewLogger(level, format string) Logger {
	log := logrus.New()
	
	// Set log level
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
	
	// Set format
	if format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
	
	// Set output
	log.SetOutput(os.Stdout)
	
	return &LogrusLogger{logger: logrus.NewEntry(log)}
}

// NewLoggerWithOutput creates a new logger with custom output
func NewLoggerWithOutput(level, format string, output io.Writer) Logger {
	log := logrus.New()
	
	// Set log level
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
	
	// Set format
	if format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
	
	// Set output
	log.SetOutput(output)
	
	return &LogrusLogger{logger: logrus.NewEntry(log)}
}

// Debug logs a debug message
func (l *LogrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// Info logs an info message
func (l *LogrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

// Warn logs a warning message
func (l *LogrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Error logs an error message
func (l *LogrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

// Fatal logs a fatal message and exits
func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Debugf logs a debug message with formatting
func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Infof logs an info message with formatting
func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warnf logs a warning message with formatting
func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Errorf logs an error message with formatting
func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Fatalf logs a fatal message with formatting and exits
func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// WithField adds a single field to the logger
func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusLogger{logger: l.logger.WithField(key, value)}
}

// WithFields adds multiple fields to the logger
func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogrusLogger{logger: l.logger.WithFields(fields)}
}
