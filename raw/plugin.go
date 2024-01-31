package raw

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Plugin creates this raw plugin
func Plugin(ctx context.Context) *plugin.Plugin {

	p := &plugin.Plugin{
		Name: "steampipe-plugin-raw",
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
		},
		DefaultTransform: transform.FromGo(),
		SchemaMode:       plugin.SchemaModeDynamic,
		TableMapFunc:     getTabels,
	}

	return p
}

// Function to get the schema dynamically
func getTabels(ctx context.Context, d *plugin.TableMapData) (map[string]*plugin.Table, error) {
	// Initialize tables
	tables := map[string]*plugin.Table{}

	plugin.Logger(ctx).Warn("############### init ######################")

	config := GetConfig(d.Connection)
	plugin.Logger(ctx).Warn("loading endpoints", "file ", *config.EndpointsFile)

	file, _ := os.Open(*config.EndpointsFile)

	var endpoints map[string]endpointConfig
	byteValue, _ := io.ReadAll(file)
	defer file.Close()
	json.Unmarshal([]byte(byteValue), &endpoints)

	for name, endpoint := range endpoints {
		columns := []*plugin.Column{}
		keyColums := []*plugin.KeyColumn{}
		plugin.Logger(ctx).Warn("init", "endpoint", name, "config ", endpoint)

		for name, fieldType := range endpoint.Fields {
			col := plugin.Column{
				Name:        name,
				Description: "",
				Type:        getType(fieldType),
				Transform:   transform.FromField(name),
			}
			columns = append(columns, &col)
		}

		if endpoint.Arguments != nil {
			for name, arg := range endpoint.Arguments {
				col := plugin.Column{
					Name:        argName(name),
					Description: "",
					Type:        getType(arg.Type),
					Transform:   transform.FromField(argName(name)),
				}

				var required string
				if arg.Optional {
					required = plugin.Optional
				} else {
					required = plugin.Required
				}

				keyCol := plugin.KeyColumn{
					Name:      argName(name),
					Operators: []string{"="},
					Require:   required,
				}
				columns = append(columns, &col)
				keyColums = append(keyColums, &keyCol)
			}
		}

		table := plugin.Table{
			Name:        endpoint.Name,
			Description: endpoint.Description,
			List: &plugin.ListConfig{
				Hydrate:    listTable,
				KeyColumns: keyColums,
			},
			Columns: columns,
		}

		tables[name] = &table
	}

	return tables, nil
}

func getType(fieldType string) proto.ColumnType {
	if fieldType == "string" {
		return proto.ColumnType_STRING
	} else if fieldType == "int" {
		return proto.ColumnType_INT
	} else if fieldType == "double" {
		return proto.ColumnType_DOUBLE
	} else if fieldType == "bool" {
		return proto.ColumnType_BOOL
	} else if fieldType == "timestamp" {
		return proto.ColumnType_TIMESTAMP
	} else if fieldType == "json" {
		return proto.ColumnType_JSON
	} else {
		panic("Unknown type" + fieldType)
	}
}

func listTable(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Warn("############### Listing ######################")
	globalConfig, _ := d.Connection.Config.(rawConfig)

	plugin.Logger(ctx).Warn("loading endpoints", "file ", *globalConfig.EndpointsFile)
	file, _ := os.Open(*globalConfig.EndpointsFile)

	var endpoints map[string]endpointConfig

	byteValue, _ := io.ReadAll(file)
	defer file.Close()
	json.Unmarshal(byteValue, &endpoints)

	endpoint := endpoints[d.Table.Name]

	plugin.Logger(ctx).Warn("config", endpoint)

	urlArgs := url.Values{}
	args := map[string]string{}
	for name, arg := range endpoint.Arguments {
		_, exists := d.EqualsQuals[argName(name)]

		if !arg.Optional && !exists {
			panic("Required argument " + argName(name) + " not provided")
		} else if exists {
			value := d.EqualsQualString(argName(name))
			urlArgs.Add(name, value)
			args[argName(name)] = value
			plugin.Logger(ctx).Warn("Added arg", "name", name, "value", value)
		}

	}

	var url string
	if len(urlArgs) > 0 {
		url = endpoint.Url + "?" + urlArgs.Encode()
	} else {
		url = endpoint.Url
	}
	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		panic(err)
	}

	for _, item := range data {
		for name, value := range args {
			item[name] = value
		}
		d.StreamListItem(ctx, item)
		plugin.Logger(ctx).Warn("Streaming", "item", item)

		// Context can be cancelled due to manual cancellation or the limit has been hit
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func argName(name string) string {
	return "_" + name
}
