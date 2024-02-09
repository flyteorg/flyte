package gormimpl

const (
	orgColumn = "org"
)

func getOrgFilter(org string) map[string]interface{} {
	return map[string]interface{}{orgColumn: org}
}
