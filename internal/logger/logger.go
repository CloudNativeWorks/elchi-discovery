package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

type Config struct {
	Level  string
	Format string
	Output string
}

func New(cfg *Config) *Logger {
	log := logrus.New()

	// Set log level
	level := logrus.InfoLevel
	if cfg != nil && cfg.Level != "" {
		if parsedLevel, err := logrus.ParseLevel(strings.ToLower(cfg.Level)); err == nil {
			level = parsedLevel
		}
	}
	log.SetLevel(level)

	// Set formatter
	if cfg != nil && cfg.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			PadLevelText:    true,
		})
	}

	// Set output
	if cfg != nil && cfg.Output == "stderr" {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(os.Stdout)
	}

	return &Logger{Logger: log}
}

func NewDefault() *Logger {
	return New(&Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})
}

func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

func (l *Logger) WithPlugin(pluginName string) *logrus.Entry {
	return l.Logger.WithField("plugin", pluginName)
}

func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.Logger.WithField("component", component)
}
