package log

import (
	"context"
	"io"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config is log config .
type Config struct {
	LogPath   string // 日志存放路径
	AppName   string // 应用名称
	Debug     bool   // 是否开启Debug模式
	MultiFile bool   // 多文件模式根据日志级别生成文件
}

var globalLogger atomic.Pointer[zap.SugaredLogger]

func init() {
	l := zap.New(zapcore.NewTee(
		zapcore.NewCore(encoder, consoleDebugging, lowPriority)),
		zap.AddCaller(),
		zap.Development(),
		zap.AddCallerSkip(1),
	).Sugar()
	globalLogger.Store(l)
}

// Init initialize a log config .
func Init(c *Config) {

	var cores []zapcore.Core
	if c.Debug {
		cores = append(cores, zapcore.NewCore(encoder, consoleDebugging, lowPriority))
	} else {
		cores = append(cores, zapcore.NewCore(encoder, consoleDebugging, highPriority))
	}
	if c.LogPath != "" {
		if c.MultiFile {
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(filepath.Join(c.LogPath, c.AppName+"_info.log"))), infoLevel))
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(filepath.Join(c.LogPath, c.AppName+"_warn.log"))), warnLevel))
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(filepath.Join(c.LogPath, c.AppName+"_error.log"))), errorLevel))
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(filepath.Join(c.LogPath, c.AppName+"_fatal.log"))), fatalLevel))
			if c.Debug {
				cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(filepath.Join(c.LogPath, c.AppName+"_debug.log"))), debugLevel))
			}
		} else {
			if c.Debug {
				cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(filepath.Join(c.LogPath, c.AppName+".log"))), lowPriority))
			} else {
				cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(filepath.Join(c.LogPath, c.AppName+".log"))), highPriority))
			}
		}
	}
	setLogger(zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.Development(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.WarnLevel),
	).Sugar())
}

type contextKey int

const traceIDKey contextKey = iota

// WithTraceID set traceID into context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// FromContext get traceID from context
func FromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(traceIDKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// TimeEncoder time encoder .
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(strconv.FormatInt(t.Unix(), 10))
}

// Debugf log
func Debugf(msg string, args ...interface{}) {
	logger().Debugf(msg, args...)
}

// Debug log
func Debug(args ...interface{}) {
	logger().Debug(args...)
}

// Infof log
func Infof(msg string, args ...interface{}) {
	logger().Infof(msg, args...)
}

// Info log
func Info(args ...interface{}) {
	logger().Info(args...)
}

// Errorf log
func Errorf(msg string, args ...interface{}) {
	logger().Errorf(msg, args...)
}

// Error log
func Error(args ...interface{}) {
	logger().Error(args...)
}

// Warnf log
func Warnf(msg string, args ...interface{}) {
	logger().Warnf(msg, args...)
}

// Warn log
func Warn(args ...interface{}) {
	logger().Warn(args...)
}

// Fatalf send log fatalf
func Fatalf(msg string, args ...interface{}) {
	logger().Fatalf(msg, args...)
}

// Fatal send log fatal
func Fatal(args ...interface{}) {
	logger().Fatal(args...)
}

func withCtx(ctx context.Context, l *zap.SugaredLogger) *zap.SugaredLogger {
	if ctx == nil {
		return l
	}
	if tid, ok := FromContext(ctx); ok {
		return l.With("trace_id", tid)
	}
	return l
}

// CtxDebug log with context
func CtxDebug(ctx context.Context, args ...interface{}) {
	withCtx(ctx, logger()).Debug(args...)
}

// CtxDebugf log with context
func CtxDebugf(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx, logger()).Debugf(msg, args...)
}

// CtxInfo log with context
func CtxInfo(ctx context.Context, args ...interface{}) {
	withCtx(ctx, logger()).Info(args...)
}

// CtxInfof log with context
func CtxInfof(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx, logger()).Infof(msg, args...)
}

// CtxWarn log with context
func CtxWarn(ctx context.Context, args ...interface{}) {
	withCtx(ctx, logger()).Warn(args...)
}

// CtxWarnf log with context
func CtxWarnf(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx, logger()).Warnf(msg, args...)
}

// CtxError log with context
func CtxError(ctx context.Context, args ...interface{}) {
	withCtx(ctx, logger()).Error(args...)
}

// CtxErrorf log with context
func CtxErrorf(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx, logger()).Errorf(msg, args...)
}

func getWriter(filename string) io.Writer {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100,  // 单个日志文件最大 100 MB
		MaxAge:     7,    // 默认保存时间为 7 天
		MaxBackups: 30,   // 最多保留 30 个备份文件
		Compress:   true, // 开启压缩
	}
}

func logger() *zap.SugaredLogger {
	return globalLogger.Load()
}

func setLogger(l *zap.SugaredLogger) {
	globalLogger.Store(l)
}
