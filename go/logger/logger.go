// from github.com/amitrai48/logger/logger.go

package logger

import "errors"

// A global variable so that log functions can be directly accessed
var log Logger

//Fields Type to pass when we want to call WithFields for structured logging
type Fields map[string]interface{}

const (
	//Debug has verbose message
	Debug = "debug"
	//Info is default log level
	Info = "info"
	//Warn is for logging messages about possible issues
	Warn = "warn"
	//Error is for logging errors
	Error = "error"
	//Fatal is for logging fatal messages. The sytem shutsdown after logging the message.
	Fatal = "fatal"
)

const (
	InstanceZapLogger int = iota
	InstanceLogrusLogger
)

type FormatType int

const (
	TypeJSONFormat FormatType = iota
	TypeTextFormat
	TypeCEFormat
)

var (
	errInvalidLoggerInstance = errors.New("Invalid logger instance")
)

//Logger is our contract for the logger
type Logger interface {
	Print(args ...interface{})

	Printf(format string, args ...interface{})

	Println(args ...interface{})

	Debugf(format string, args ...interface{})

	Infof(format string, args ...interface{})

	Warnf(format string, args ...interface{})

	Errorf(format string, args ...interface{})

	Fatalf(format string, args ...interface{})

	Panicf(format string, args ...interface{})

	WithFields(keyValues Fields) Logger
}

// Configuration stores the config for the logger
// For some loggers there can only be one level across writers, for such the level of Console is picked by default
type Configuration struct {
	LogLevel             string
	EnableCloudEvents    bool
	EnableKafka          bool
	KafkaFormat          FormatType
	KafkaProducerCfg     ProducerConfiguration
	EnableConsole        bool
	ConsoleFormat        FormatType
	ConsoleLevelTruncate bool
	EnableFile           bool
	FileFormat           FormatType
	FileLevelTruncate    bool
	FileLocation         string
}

//NewLogger returns an instance of logger
func NewLogger(config Configuration, loggerInstance int) (Logger, error) {
	switch loggerInstance {
	case InstanceZapLogger:
		logger, err := newZapLogger(config)
		if err != nil {
			return nil, err
		}
		log = logger
		return logger, nil

	case InstanceLogrusLogger:
		logger, err := newLogrusLogger(config)
		if err != nil {
			return nil, err
		}
		log = logger
		return logger, nil

	default:
		return nil, errInvalidLoggerInstance
	}
}

func Print(args ...interface{}) {
	log.Print(args...)
}

func Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func Println(args ...interface{}) {
	log.Println(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

func WithFields(keyValues Fields) Logger {
	return log.WithFields(keyValues)
}
