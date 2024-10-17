package zap_log

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/errors"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func (p *provider) Logger(_ contracts.Application) logger.Logger {
	return p.log
}

func (p *provider) getZapLogger(isDev bool, inCfg *zap.Config) *zap.Logger {
	if p.zap != nil {
		return p.zap
	}

	loadedConfig, err := p.loadConfig()
	if err != nil {
		panic(errors.Errorf("failed to load log config: %s", err.Error()))
	}

	var cfg zap.Config
	if isDev {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// конфиг, который указан в провайдере методом WithConfig
	if p.config != nil {
		cfg = mergeConfig(cfg, *p.config)
	}
	// конфиг, который зарегистрирован в контейнере
	if inCfg != nil {
		cfg = mergeConfig(cfg, *inCfg)
	}
	// конфиг, который загружен из файла
	if loadedConfig != nil {
		cfg = mergeConfig(cfg, *loadedConfig)
	}

	if p.level != "" {
		var lvl zapcore.Level
		err := lvl.Set(p.level)
		if err == nil {
			cfg.Level = zap.NewAtomicLevelAt(lvl)
		}
	}

	cfg.Development = isDev
	cfg.Encoding = "json"
	cfg.DisableStacktrace = true

	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	p.zap = log

	return p.zap
}

func (p *provider) loadConfig() (*zap.Config, error) {
	if p.cfgFile == "" {
		return nil, nil
	}
	cfg := &config{}
	file, err := os.Open(p.cfgFile)
	if err != nil {
		return nil, err
	}

	fileBody, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(fileBody, cfg)
	if err != nil {
		return nil, err
	}

	return cfg.Log, nil
}

func mergeConfig(target, source zap.Config) zap.Config {
	if source.Level != emptyLevel {
		target.Level = source.Level
	}
	target.Development = source.Development
	target.DisableCaller = source.DisableCaller
	target.DisableStacktrace = source.DisableStacktrace
	target.Encoding = source.Encoding
	if source.Sampling != nil {
		target.Sampling = source.Sampling
	}
	if source.OutputPaths != nil {
		target.OutputPaths = source.OutputPaths
	}
	if source.ErrorOutputPaths != nil {
		target.ErrorOutputPaths = source.ErrorOutputPaths
	}
	if source.InitialFields != nil {
		target.InitialFields = source.InitialFields
	}
	target.EncoderConfig = mergeEncoderConfig(target.EncoderConfig, source.EncoderConfig)
	return target
}

func mergeEncoderConfig(target, source zapcore.EncoderConfig) zapcore.EncoderConfig {
	if source.MessageKey != "" {
		target.MessageKey = source.MessageKey
	}
	if source.LevelKey != "" {
		target.LevelKey = source.LevelKey
	}
	if source.TimeKey != "" {
		target.TimeKey = source.TimeKey
	}
	if source.NameKey != "" {
		target.NameKey = source.NameKey
	}
	if source.CallerKey != "" {
		target.CallerKey = source.CallerKey
	}
	if source.FunctionKey != "" {
		target.FunctionKey = source.FunctionKey
	}
	if source.StacktraceKey != "" {
		target.StacktraceKey = source.StacktraceKey
	}
	if source.LineEnding != "" {
		target.LineEnding = source.LineEnding
	}
	if source.EncodeLevel != nil {
		target.EncodeLevel = source.EncodeLevel
	}
	if source.EncodeTime != nil {
		target.EncodeTime = source.EncodeTime
	}
	if source.EncodeDuration != nil {
		target.EncodeDuration = source.EncodeDuration
	}
	if source.EncodeCaller != nil {
		target.EncodeCaller = source.EncodeCaller
	}
	if source.EncodeName != nil {
		target.EncodeName = source.EncodeName
	}
	if source.ConsoleSeparator != "" {
		target.ConsoleSeparator = source.ConsoleSeparator
	}
	return target
}
