// from github.com/amitrai48/logger/zap.go

package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
}

func getEncoder(format FormatType) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	switch format {
	case TypeJSONFormat:
		return zapcore.NewJSONEncoder(encoderConfig)
	case TypeTextFormat:
		return zapcore.NewConsoleEncoder(encoderConfig)
	case TypeCEFormat:
		encoderConfig.TimeKey = ceTimeKey
		encoderConfig.LevelKey = ceLevelKey
		encoderConfig.MessageKey = ceMessageKey
		encoderConfig.CallerKey = ""
		return zapcore.NewJSONEncoder(encoderConfig)
	default:
		return nil
	}
}

func getZapLevel(level string) zapcore.Level {
	switch level {
	case Info:
		return zapcore.InfoLevel
	case Warn:
		return zapcore.WarnLevel
	case Debug:
		return zapcore.DebugLevel
	case Error:
		return zapcore.ErrorLevel
	case Fatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func newZapLogger(config Configuration) (Logger, error) {
	cores := []zapcore.Core{}

	if config.EnableKafka {
		level := getZapLevel(config.LogLevel)
		// create an async producer
		asyncproducer, err := NewAsyncProducer(config.KafkaProducerCfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "NewAsyncProducer failed", err.Error())
		}
		writer := NewZapWriter(config.KafkaProducerCfg.Topic, asyncproducer)
		core := zapcore.NewCore(getEncoder(config.KafkaFormat), writer, level)
		cores = append(cores, core)
	}

	if config.EnableConsole {
		level := getZapLevel(config.LogLevel)
		writer := zapcore.Lock(os.Stdout)
		core := zapcore.NewCore(getEncoder(config.ConsoleFormat), writer, level)
		cores = append(cores, core)
	}

	if config.EnableFile {
		level := getZapLevel(config.LogLevel)
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename: config.FileLocation,
			MaxSize:  100,
			Compress: true,
			MaxAge:   28,
		})
		core := zapcore.NewCore(getEncoder(config.FileFormat), writer, level)
		cores = append(cores, core)
	}

	combinedCore := zapcore.NewTee(cores...)
	logger := zap.New(combinedCore).Sugar()

	if config.EnableCloudEvents {
		zaplogger := &zapLogger{
			sugaredLogger: logger,
		}
		return zaplogger.WithFields(ceFields), nil
	}

	return &zapLogger{
		sugaredLogger: logger,
	}, nil
}

func (l *zapLogger) Print(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Printf(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Println(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.sugaredLogger.Debugf(format, args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.sugaredLogger.Warnf(format, args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.sugaredLogger.Errorf(format, args...)
}

func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) Panicf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) WithFields(fields Fields) Logger {
	var f = make([]interface{}, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}
	newLogger := l.sugaredLogger.With(f...)
	return &zapLogger{newLogger}
}
