package ch13

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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
	// {"level":"debug","name":"example","caller":"ch13-logging-and-metrics/zap_test.go:63","msg":"test debug message","version":"go1.23.4"}
	// {"level":"info","name":"example","caller":"ch13-logging-and-metrics/zap_test.go:65","msg":"test info message","version":"go1.23.4"}
}

func Example_zapConsole() {
	// create ecoder, sink, and core (info level)
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	syncer := zapcore.Lock(os.Stdout)
	core := zapcore.NewCore(encoder, syncer, zapcore.InfoLevel)

	// create logger (options are empty meaning caller won't be logged)
	zl := zap.New(core)
	// sync at scope exit
	defer func() {
		_ = zl.Sync()
	}()

	// create named console logger
	console := zl.Named("[console]")
	// log with levels info, debug, and error
	console.Info("this is logged by the logger")
	console.Debug("this is below the logger's threshold and won't log")
	console.Error("this is also logged by the logger")

	// Output:
	// info	[console]	this is logged by the logger
	// error	[console]	this is also logged by the logger
}

// func Example_zapInfoFileDebugConsole
func Example_zapInfoFileDebugConsole() {
	// mock a log file
	logFile := new(bytes.Buffer)

	// default core: json encoder, sync to file, level = info
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	syncer := zapcore.AddSync(logFile)
	core := zapcore.NewCore(encoder, syncer, zapcore.InfoLevel)

	// deafult loger from the default core
	zl := zap.New(core)
	defer func() {
		_ = zl.Sync()
	}()

	// log some initial messages: debug (shouldn't be displayed) + error
	zl.Debug("this is below the logger's threshold and won't log")
	zl.Error("this is logged by the logger")

	// core wrapper, to "branch" additional logger
	coreWrapper := func(core zapcore.Core) zapcore.Core {
		// copy encoder config to modify a copy
		configCopy := encoderConfig
		// capitalize log level as an example modification
		configCopy.EncodeLevel = zapcore.CapitalLevelEncoder

		// new console encoder, console syncer, debug-level console
		encoder := zapcore.NewConsoleEncoder(configCopy)
		syncer := zapcore.Lock(os.Stdout)
		consoleCore := zapcore.NewCore(encoder, syncer, zapcore.DebugLevel)

		// merge wrapped core and new core
		cores := []zapcore.Core{
			core,
			consoleCore,
		}

		// create tee to log to the two cores at the same time
		return zapcore.NewTee(cores...)
	}

	// specify options for the new logger
	options := []zap.Option{
		zap.WrapCore(coreWrapper),
	}

	// modify logger with options
	// (at this point we start logging to both console and file)
	zl = zl.WithOptions(options...)

	// log header
	fmt.Println("standard output:")
	// log debug + info
	zl.Debug("this is only logged as console encoding")
	zl.Info("this is logged as console encoding and JSON")

	// print log file contents
	fmt.Print("\nfile contents:\n", logFile.String())

	// Output:
	// standard output:
	// DEBUG	this is only logged as console encoding
	// INFO	this is logged as console encoding and JSON
	//
	// file contents:
	// {"level":"error","msg":"this is logged by the logger"}
	// {"level":"info","msg":"this is logged as console encoding and JSON"}
}

// func Example_zapSampling
func Example_zapSampling() {
	// new core (format: json, syncer: stdout)
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	syncer := zapcore.Lock(os.Stdout)
	core := zapcore.NewCore(encoder, syncer, zapcore.DebugLevel)

	// new sampler (tick = 1 sec, first = 1, thereafter = 3)
	sampler := zapcore.NewSamplerWithOptions(core, time.Second, 1, 3)

	// new log based on the sampler
	zl := zap.New(sampler)
	// defer sync
	defer func() {
		_ = zl.Sync()
	}()

	// iterate 10 times, the sampler should log each unique message and
	// only each 3rd repetetive one or when sampling interval is expired
	for i := 0; i < 10; i++ {
		// sleep on the 5th iteration to "force" the sampler to log
		if i == 5 {
			time.Sleep(time.Second)
		}

		// unique debug message
		zl.Debug(fmt.Sprintf("%d", i))
		// repetetive debug message
		zl.Debug("the same message")
	}

	// Output:
	// {"level":"debug","msg":"0"}
	// {"level":"debug","msg":"the same message"}
	// {"level":"debug","msg":"1"}
	// {"level":"debug","msg":"2"}
	// {"level":"debug","msg":"3"}
	// {"level":"debug","msg":"the same message"}
	// {"level":"debug","msg":"4"}
	// {"level":"debug","msg":"5"}
	// {"level":"debug","msg":"the same message"}
	// {"level":"debug","msg":"6"}
	// {"level":"debug","msg":"7"}
	// {"level":"debug","msg":"8"}
	// {"level":"debug","msg":"the same message"}
	// {"level":"debug","msg":"9"}
}

// func Example_zapDynamicDebugging
func Example_zapDynamicDebugging() {
	// create a JSON-encoder, syncer, but atomic leveler instead of a constant value
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	syncer := zapcore.Lock(os.Stdout)
	atomicLevel := zap.NewAtomicLevel()
	// create the core
	core := zapcore.NewCore(encoder, syncer, atomicLevel)

	// create the logger
	zl := zap.New(core)
	defer func() {
		_ = zl.Sync()
	}()

	// make a temporary directory for the semaphore file
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()
	// semaphore file name
	semaphoreFile := filepath.Join(tempDir, "level.debug")

	// run semaphore file state change tracking in the background
	ready := make(chan struct{})
	go changeLogLevelOnFileChange(atomicLevel, semaphoreFile, ready)
	// wait for the goroutine to init its state
	<-ready

	// try logging a debug message (shouldn't log)
	zl.Debug("this is below the logger's threshold")

	// create the semaphore file and close it immediately (since the existence is only important)
	sf, err := os.Create(semaphoreFile)
	if err != nil {
		log.Fatal(err)
	}
	_ = sf.Close()
	// wait for the goroutine to catch up with the filesystem changes
	<-ready

	// try logging a debug message (should log this time)
	zl.Debug("this is now at the logger's threshold")

	// remove the semaphore file and wait for the goroutine to handle it
	err = os.Remove(semaphoreFile)
	if err != nil {
		log.Fatal(err)
	}
	<-ready

	// try logging a debug and an info messages (should log info only)
	zl.Debug("this is below the logger's threshold again")
	zl.Info("this is at the logger's current threshold")

	// Output:
	// {"level":"debug","msg":"this is now at the logger's threshold"}
	// {"level":"info","msg":"this is at the logger's current threshold"}
}

// func changeLogLevelOnFileChange
func changeLogLevelOnFileChange(
	atomicLevel zap.AtomicLevel,
	semaphoreFile string,
	ready chan struct{},
) {
	// create a filesystem watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	// close the watcher at scope exit
	defer func() {
		_ = watcher.Close()
	}()

	// add the semaphore file's directory to the watched list
	dir := filepath.Dir(semaphoreFile)
	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	// preserve the original level before changes
	originalLevel := atomicLevel.Level()

	// notify ready to track changes
	ready <- struct{}{}

	// run in a loop
	for {
		// wait for either a watcher event or an error
		// in both cases, return if a channel is closed
		select {
		case e, ok := <-watcher.Events:
			if !ok {
				return
			}
			// filter only the events related to the semaphore file
			if e.Name == semaphoreFile {
				switch {
				case e.Op&fsnotify.Create == fsnotify.Create:
					// if the operation is create, set debug level to debug
					atomicLevel.SetLevel(zapcore.DebugLevel)
					ready <- struct{}{}

				case e.Op&fsnotify.Remove == fsnotify.Remove:
					// if the operation is remove, set back the original level
					atomicLevel.SetLevel(originalLevel)
					ready <- struct{}{}
				}
			}
		case e, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Fatal(e.Error())
		}
	}
}

// func Example_zapLogRotation
func Example_zapLogRotation() {
	// create a temporary directory and clean it up at scope exit
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// create a typical JSON encoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	// create a lumberjack logger managing log file rotation etc.
	rotatingLogger := &lumberjack.Logger{
		// specify filename (default is <processname>-lumberjack.log)
		Filename: filepath.Join(tempDir, "debug.log"),
		// maximum size of a log file in megabytes
		MaxSize: 100,
		// maximum logfile age before it should be rotated
		MaxAge: 7,
		// maximum amount of rotated log files to keep
		MaxBackups: 5,
		// use local time instead of UTC
		LocalTime: true,
		// do compress
		Compress: true,
	}
	// add output synchronization to the lumberjack logger
	syncer := zapcore.AddSync(rotatingLogger)
	// create a core
	core := zapcore.NewCore(encoder, syncer, zapcore.DebugLevel)

	// create a logger
	zl := zap.New(core)
	defer func() {
		_ = zl.Sync()
	}()

	// log some message
	zl.Debug("debug messsage written to the log file")

	// Output:
}
