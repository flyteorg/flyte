package filters

import (
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func BuildResourceListRequestWithName(c Filters, project, domain, name string) (*admin.ResourceListRequest, error) {
	fieldSelector, err := Transform(SplitTerms(c.FieldSelector))
	if err != nil {
		return nil, err
	}
	request := &admin.ResourceListRequest{
		Limit:   uint32(c.Limit),
		Token:   getToken(c),
		Filters: fieldSelector,
		Id: &admin.NamedEntityIdentifier{
			Project: project,
			Domain:  domain,
		},
	}
	if len(name) > 0 {
		request.Id.Name = name
	}
	if sort := buildSortingRequest(c); sort != nil {
		request.SortBy = sort
	}
	return request, nil
}

func BuildProjectListRequest(c Filters) (*admin.ProjectListRequest, error) {
	fieldSelector, err := Transform(SplitTerms(c.FieldSelector))
	if err != nil {
		return nil, err
	}
	request := &admin.ProjectListRequest{
		Limit:   uint32(c.Limit),
		Token:   getToken(c),
		Filters: fieldSelector,
		SortBy:  buildSortingRequest(c),
	}
	return request, nil
}

func buildSortingRequest(c Filters) *admin.Sort {
	sortingOrder := admin.Sort_DESCENDING
	if c.Asc {
		sortingOrder = admin.Sort_ASCENDING
	}
	if len(c.SortBy) > 0 {
		return &admin.Sort{
			Key:       c.SortBy,
			Direction: sortingOrder,
		}
	}
	return nil
}

func getToken(c Filters) string {
	token := int(c.Page-1) * int(c.Limit)
	if token <= 0 {
		return ""
	}
	return strconv.Itoa(token)
}
