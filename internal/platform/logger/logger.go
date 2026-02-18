package logger

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/imidll/selfshop.ms-catalog/internal/platform/config"
)

const (
	samplerInitial    = 100
	samplerThereafter = 10
	samplerInterval   = time.Second
)

type T struct {
	*zap.Logger
	unsampled *zap.Logger
	level     zap.AtomicLevel
}

func Transform(logr *T) fxevent.Logger {
	return &fxevent.ZapLogger{
		Logger: logr.unsampled.WithOptions(
			zap.IncreaseLevel(zap.WarnLevel),
			zap.AddCallerSkip(-1)),
	}
}

func MustNew(conf *config.T) *T {
	l, err := New(conf)
	if err != nil {
		panic(err)
	}
	return l
}

func New(conf *config.T) (*T, error) {
	atomicLevel, err := zap.ParseAtomicLevel(conf.Log.MinLevel)
	if err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", conf.Log.MinLevel, err)
	}

	encConfig := zap.NewProductionEncoderConfig()
	encConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encConfig.EncodeDuration = zapcore.MillisDurationEncoder
	encConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var encoder zapcore.Encoder
	switch conf.GetLogFormat() {
	case config.LogFormatJSON:
		encoder = zapcore.NewJSONEncoder(encConfig)
	default: // console, text, pretty, ...
		encConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encConfig)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), atomicLevel)

	sampledCore := zapcore.NewSamplerWithOptions(
		core,
		samplerInterval,
		samplerInitial,
		samplerThereafter,
	)

	opts := []zap.Option{
		zap.AddCaller(),
		zap.ErrorOutput(zapcore.AddSync(os.Stderr)),
		zap.Fields(
			zap.String("appname", conf.App.Name),
			zap.String("runmode", conf.App.Runmode),
		),
	}

	return &T{
		Logger:    zap.New(sampledCore, opts...),
		unsampled: zap.New(core, opts...),
		level:     atomicLevel,
	}, nil
}

func (l *T) Critical() *zap.Logger {
	return l.unsampled
}

func (l *T) GetLevel() zapcore.Level {
	return l.level.Level()
}

func (l *T) SetLevel(s string) error {
	if s == "" {
		return errors.New("log level cannot be empty")
	}

	lvl, err := zapcore.ParseLevel(s)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", s, err)
	}

	prev := l.level.Level()

	if lvl == prev {
		l.unsampled.Debug("log level already set", zap.String("level", prev.String()))
		return nil
	}

	l.unsampled.Info("log level changed",
		zap.String("from", prev.String()),
		zap.String("to", lvl.String()),
	)

	l.level.SetLevel(lvl)
	return nil
}

func (l *T) Named(title string) *T {
	if title == "" {
		return l
	}
	return &T{
		Logger:    l.Logger.Named(title),
		unsampled: l.unsampled.Named(title),
		level:     l.level,
	}
}

func (l *T) With(fields ...zap.Field) *T {
	if len(fields) == 0 {
		return l
	}
	return &T{
		Logger:    l.Logger.With(fields...),
		unsampled: l.unsampled.With(fields...),
		level:     l.level,
	}
}

func (l *T) WithOptions(options ...zap.Option) *T {
	if len(options) == 0 {
		return l
	}
	return &T{
		Logger:    l.Logger.WithOptions(options...),
		unsampled: l.unsampled.WithOptions(options...),
		level:     l.level,
	}
}

func (l *T) Sync() error {
	return errors.Join(
		l.Logger.Sync(),
		l.unsampled.Sync(),
	)
}
