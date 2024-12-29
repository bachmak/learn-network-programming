package ch13

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// encoder config (agnostic of specific formats encoders use)
var encoderConfig = zapcore.EncoderConfig{
	// message and name keys
	MessageKey: "msg",
	NameKey:    "name",

	// level keys + level format
	LevelKey:    "level",
	EncodeLevel: zapcore.LowercaseLevelEncoder,

	// caller key + caller format
	CallerKey:    "caller",
	EncodeCaller: zapcore.ShortCallerEncoder,

	// time key + time format (not used for deterministic tests)
	// TimeKey: "time",
	// EncodeTime: zapcore.ISO8601TimeEncoder,
}

// func Example_zapJSON
func Example_zapJSON() {
	// new JSON encoder from the config
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	// write syncer to stduot
	syncer := zapcore.Lock(os.Stdout)
	// new core, log level = debug
	core := zapcore.NewCore(encoder, syncer, zapcore.DebugLevel)

	// options:
	options := []zap.Option{
		// log the caller
		zap.AddCaller(),
		// add fields: go runtime version (stubbed for future test correctness)
		zap.Fields(zap.String("version", "go1.23.4")),
	}

	// new logger from the core and the options
	zl := zap.New(core, options...)
	// flush at scope exit
	defer func() {
		_ = zl.Sync()
	}()

	// create a "child" logger, named "example"
	example := zl.Named("example")
	// log debug message
	example.Debug("test debug message")
	// log info message
	example.Info("test info message")

	// Output:
	// {"level":"debug","name":"example","caller":"ch13-logging-and-metrics/zap_test.go:56","msg":"test debug message","version":"go1.23.4"}
	// {"level":"info","name":"example","caller":"ch13-logging-and-metrics/zap_test.go:58","msg":"test info message","version":"go1.23.4"}
}
