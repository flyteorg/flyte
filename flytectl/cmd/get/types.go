package get

import "encoding/json"

type PrintableNamedEntityIdentifier struct {
	Name    string `header:"Name"`
	Project string `header:"Project"`
	Domain  string `header:"Domain"`
}

var entityStructure = map[string]string{
	"Domain":  "$.domain",
	"Name":    "$.name",
	"Project": "$.project",
}

var transformTaskEntity = func(jsonbody []byte) (interface{}, error) {
	results := PrintableNamedEntityIdentifier{}
	if err := json.Unmarshal(jsonbody, &results); err != nil {
		return results, err
	}
	return results, nil
}
