package tomladapter

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/pelletier/go-toml"
)

// automatically register config adapter to caddyconfig on init
func init() {
	caddyconfig.RegisterAdapter("toml", Adapter{})
}

// struct Adapter
type Adapter struct{}

// implement Adapt method (transforms original config to caddy json)
func (Adapter) Adapt(body []byte, _ map[string]interface{}) (
	[]byte,
	[]caddyconfig.Warning,
	error,
) {
	// build a toml-tree from the content
	tree, err := toml.LoadBytes(body)
	if err != nil {
		return nil, nil, err
	}

	// serialize the toml-tree to json
	config, err := json.Marshal(tree.ToMap())
	if err != nil {
		return nil, nil, err
	}

	// return the resulting config
	return config, nil, nil
}

// interface implementation guard
var _ caddyconfig.Adapter = (*Adapter)(nil)
