package handlers

import (
	"html/template"
	"io"
	"net/http"
)

// template for writing response while escaping HTML characters
var t = template.Must(template.New("hello").Parse("Hello, {{.}}!"))

// Func DefaultHandler
func DefaultHandler() http.Handler {
	// We wrap our function into the http.Handler interface
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Always explicitly drain request body on exit and close it
			defer func(r io.ReadCloser) {
				_, _ = io.Copy(io.Discard, r)
				_ = r.Close()
			}(r.Body)

			// Do the logic
			content, hErr := defaultHandlerImpl(r.Method, r.Body)
			// If there's an error, write the error
			if hErr != nil {
				http.Error(w, hErr.errMsg, hErr.errCode)
			}

			// Safely write the message using the template
			_ = t.Execute(w, content)
		},
	)
}

// Func defaultHandlerImpl
func defaultHandlerImpl(method string, body io.Reader) (string, *handlerError) {
	// Depending on the request method
	switch method {
	// on get request, return a welcoming string
	case http.MethodGet:
		return "friend", nil
		// on post request, read the message body and use it to reply
	case http.MethodPost:
		content, err := io.ReadAll(body)
		// return internal server error on read failure
		if err != nil {
			return "", &handlerError{
				errCode: http.StatusInternalServerError,
				errMsg:  "Internval server error",
			}
		}
		return string(content), nil
	}

	// otherwise, method is not allowed
	return "", &handlerError{
		errCode: http.StatusMethodNotAllowed,
		errMsg:  "Method not allowed",
	}
}

// handlerError type
type handlerError struct {
	errCode int
	errMsg  string
}
