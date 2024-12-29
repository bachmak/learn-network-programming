package ch13

import "go.uber.org/zap/zapcore"

// encoder config (agnostic of specific formats encoders use)
var encoderConfig = zapcore.EncoderConfig{
	// message and name keys
	MessageKey: "msg",
	NameKey:    "name",

	// level keys + level format
	LevelKey:    "level",
	EncodeLevel: zapcore.LowercaseColorLevelEncoder,

	// caller key + caller format
	CallerKey:    "caller",
	EncodeCaller: zapcore.ShortCallerEncoder,

	// time key + time format (not used for deterministic tests)
	// TimeKey: "time",
	// EncodeTime: zapcore.ISO8601TimeEncoder,
}
