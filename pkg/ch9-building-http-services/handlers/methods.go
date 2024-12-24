package handlers

import (
	"io"
	"net/http"
	"sort"
	"strings"
)

// type Methods (map from a method name to handlers)
type Methods map[string]http.Handler

// Func ServeHTTP (as Methods implements http.Handler interface itself)
func (m Methods) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// drain the request body here so that the handler don't have to
	defer func(r io.ReadCloser) {
		_, _ = io.Copy(io.Discard, r)
		_ = r.Close()
	}(r.Body)

	// find a handler corresponding to the request method
	if handler, ok := m[r.Method]; ok {
		// if the handler is nil, response with internal server error and return
		if handler == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// dispatch the request to the handler
		handler.ServeHTTP(w, r)
		return
	}

	// add "Allow" header entry with the list of supported methods
	w.Header().Add("Allow", m.allowedMethods())
	// if the client didn't explicitly ask for the method list, reply with 405
	if r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Func allowedMethods
func (m Methods) allowedMethods() string {
	// concatenate sorted keys from the map
	methodList := make([]string, 0, len(m))

	for k := range m {
		methodList = append(methodList, k)
	}

	sort.Strings(methodList)
	return strings.Join(methodList, ", ")
}
