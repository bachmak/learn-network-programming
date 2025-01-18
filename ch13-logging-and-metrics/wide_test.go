package ch13

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// struct wideResponseWriter
type wideResponseWriter struct {
	// embed http response writer
	http.ResponseWriter
	// store the number of written bytes and the status
	length int
	status int
}

// override WriteHeader func
func (w *wideResponseWriter) WriteHeader(statusCode int) {
	// redirect to the ResponseWriter
	w.ResponseWriter.WriteHeader(statusCode)
	// store status
	w.status = statusCode
}

// override Write func
func (w *wideResponseWriter) Write(b []byte) (int, error) {
	// redirect to the ResponseWriter
	n, err := w.ResponseWriter.Write(b)
	if err == nil {
		// accumulate length
		w.length += n
	}

	// store status OK if not set
	// (as it's typically done when body is written without status)
	if w.status == 0 {
		w.status = http.StatusOK
	}

	return n, err
}

// func WideEventLog
func WideEventLog(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(
		// wrap the next handler in a function
		func(w http.ResponseWriter, r *http.Request) {
			// create a wide writer to capture additional info
			wideWriter := &wideResponseWriter{ResponseWriter: w}

			// serve using the internal handler but the extended writer
			next.ServeHTTP(wideWriter, r)

			// get only host (for tests' predictability)
			host, _, _ := net.SplitHostPort(r.RemoteAddr)
			// log wide log entry to the zap logger
			logger.Info(
				// message
				"example wide event",
				// status code
				zap.Int("status_code", wideWriter.status),
				// response length
				zap.Int("response_length", wideWriter.length),
				// request content length
				zap.Int64("content_length", r.ContentLength),
				// request method
				zap.String("method", r.Method),
				// request protocol
				zap.String("proto", r.Proto),
				// client remote address
				zap.String("remote_addr", host),
				// request URI
				zap.String("uri", r.RequestURI),
				// request user-agent
				zap.String("user_agent", r.UserAgent()),
			)
		},
	)
}

// func Example_wideLogEntry
func Example_wideLogEntry() {
	// create a core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		zapcore.DebugLevel,
	)

	// create a logger
	zl := zap.New(core)
	defer func() {
		_ = zl.Sync()
	}()

	// init an HTTP handler
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// explicitly drain the request body
			defer func() {
				_, _ = io.Copy(io.Discard, r.Body)
				_ = r.Body.Close()
			}()

			// write hello
			w.Write([]byte("Hello!"))
		},
	)

	// create a simple HTTP server, wrapping the handler in the wide-logger-aware middleware
	ts := httptest.NewServer(WideEventLog(zl, handler))
	// close server at scope exit
	defer func() {
		ts.Close()
	}()

	// make a GET request to the server, URI = /test
	resp, err := http.Get(ts.URL + "/test")
	if err != nil {
		zl.Fatal(err.Error())
	}
	// close response body at scope exit
	defer func() {
		_ = resp.Body.Close()
	}()

	// Output:
	// {"level":"info","msg":"example wide event","status_code":200,"response_length":6,"content_length":0,"method":"GET","proto":"HTTP/1.1","remote_addr":"127.0.0.1","uri":"/test","user_agent":"Go-http-client/1.1"}
}
