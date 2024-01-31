package raw

import (
	// "fmt"
	// "strings"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type rawConfig struct {
	EndpointsFile       *string  `hcl:"endpoints_file"`
}

func ConfigInstance() interface{} {
	return &rawConfig{}
}

// GetConfig :: retrieve and cast connection config from query data
func GetConfig(connection *plugin.Connection) rawConfig {
	if connection == nil || connection.Config == nil {
		return rawConfig{}
	}
	config, _ := connection.Config.(rawConfig)

	return config
}
