package filters

var (
	DefaultLimit  int32 = 100
	DefaultFilter       = Filters{
		Limit:  DefaultLimit,
		Page:   1,
		SortBy: "created_at",
		Asc:    false,
	}
)

type Filters struct {
	FieldSelector string `json:"fieldSelector" pflag:",Specifies the Field selector"`
	SortBy        string `json:"sortBy" pflag:",Specifies which field to sort results "`
	Limit         int32  `json:"limit" pflag:",Specifies the limit"`
	Asc           bool   `json:"asc"  pflag:",Specifies the sorting order. By default flytectl sort result in descending order"`
	Page          int32  `json:"page" pflag:",Specifies the page number, in case there are multiple pages of results"`
}
