package raw

type endpointConfig struct {
	Name		string   `json:"name"`
	Description	string   `json:"description"`
	Url			string   `json:"url"`
	Headers     map[string]string `json:"headers"`
	Fields		map[string]string `json:"fields"`
	Arguments	map[string]endpointArgument `json:"arguments"`
}

type endpointArgument struct {
	Type		string   `json:"type"`
	Optional	bool     `json:"optional"`
}