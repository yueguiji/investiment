package logger

import (
	"fmt"
	"go-stock/backend/runtimepaths"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger
var SugaredLogger *zap.SugaredLogger

func init() {
	InitLogger()
}

func InitLogger() {
	_ = os.MkdirAll(runtimepaths.LogsDir(), os.ModePerm)

	encoder := getEncoder()

	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev < zap.ErrorLevel && lev >= zap.DebugLevel
	})

	infoFileWriteSyncer := getInfoWriterSyncer()
	errorFileWriteSyncer := getErrorWriterSyncer()

	infoFileCore := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(infoFileWriteSyncer, zapcore.AddSync(os.Stdout)),
		lowPriority,
	)
	errorFileCore := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(errorFileWriteSyncer, zapcore.AddSync(os.Stdout)),
		highPriority,
	)

	Logger = zap.New(zapcore.NewTee(infoFileCore, errorFileCore), zap.AddCaller())
	SugaredLogger = Logger.Sugar()
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func levelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var level string
	switch l {
	case zapcore.DebugLevel:
		level = "[DEBUG]"
	case zapcore.InfoLevel:
		level = "[INFO]"
	case zapcore.WarnLevel:
		level = "[WARN]"
	case zapcore.ErrorLevel:
		level = "[ERROR]"
	case zapcore.DPanicLevel:
		level = "[DPANIC]"
	case zapcore.PanicLevel:
		level = "[PANIC]"
	case zapcore.FatalLevel:
		level = "[FATAL]"
	default:
		level = fmt.Sprintf("[LEVEL(%d)]", l)
	}
	enc.AppendString(level)
}

func shortCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]", caller.TrimmedPath()))
}

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    levelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   shortCallerEncoder,
	}
}

func getEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(NewEncoderConfig())
}

func getInfoWriterSyncer() zapcore.WriteSyncer {
	infoLumberIO := &lumberjack.Logger{
		Filename:   filepath.Join(runtimepaths.LogsDir(), "info.log"),
		MaxSize:    10,
		MaxBackups: 100,
		MaxAge:     28,
		Compress:   false,
	}
	return zapcore.AddSync(infoLumberIO)
}

func getErrorWriterSyncer() zapcore.WriteSyncer {
	lumberWriteSyncer := &lumberjack.Logger{
		Filename:   filepath.Join(runtimepaths.LogsDir(), "error.log"),
		MaxSize:    10,
		MaxBackups: 100,
		MaxAge:     28,
		Compress:   false,
	}
	return zapcore.AddSync(lumberWriteSyncer)
}
