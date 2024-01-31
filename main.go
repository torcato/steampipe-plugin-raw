package main

import (
	"github.com/turbot/steampipe-plugin-raw/raw"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: raw.Plugin})
}
