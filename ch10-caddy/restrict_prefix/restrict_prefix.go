package restrictprefix

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

// register middleware handler as caddy module on init
func init() {
	caddy.RegisterModule(RestrictPrefix{})
}

// struct RestrictPrefix
type RestrictPrefix struct {
	// prefix + annotations for json unmarschalling
	Prefix string `json:"prefix,omitempty"`
	// logger passed from caddy
	logger *zap.Logger
}

// func CaddyModule to register the module
func (RestrictPrefix) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.restrict_prefix",
		New: func() caddy.Module { return new(RestrictPrefix) },
	}
}

// func Provision to do addional setup after loading
func (p *RestrictPrefix) Provision(ctx caddy.Context) error {
	// initialize prefix
	p.Prefix = "."
	// retrieve logger created by caddy and associaated with the module
	p.logger = ctx.Logger(p)
	return nil
}

// func Validate to validate module state after Provision is called
func (p *RestrictPrefix) Validate() error {
	// check prefix
	if p.Prefix == "" {
		return fmt.Errorf("Prefix not initialized")
	}

	// check logger
	if p.logger == nil {
		return fmt.Errorf("Logger not initialized")
	}

	return nil
}

// func ServeHTTP to implement MiddlewareHandler logic
func (p RestrictPrefix) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
	next caddyhttp.Handler,
) error {
	// iterate over the path
	for _, part := range strings.Split(r.URL.Path, "/") {
		// reply with 404 if it contains prefixed parts
		if strings.HasPrefix(part, p.Prefix) {
			http.Error(w, "Not Found", http.StatusNotFound)
			// log that access was restricted
			p.logger.Debug(
				"access restricted by prefix",
				zap.String("part", part),
				zap.String("path", r.URL.Path),
			)

			return nil
		}
	}

	// pass execution to the next handler
	return next.ServeHTTP(w, r)
}

// interface guard
var (
	_ caddy.Module                = (*RestrictPrefix)(nil)
	_ caddy.Provisioner           = (*RestrictPrefix)(nil)
	_ caddy.Validator             = (*RestrictPrefix)(nil)
	_ caddyhttp.MiddlewareHandler = (*RestrictPrefix)(nil)
)
