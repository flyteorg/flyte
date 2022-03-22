/*
 * flyteidl/service/admin.proto
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: version not set
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package flyteadmin

// Defines a set of constraints used to select eligible objects based on labels they possess.
type AdminAffinity struct {
	// Multiples selectors are 'and'-ed together to produce the list of matching, eligible objects.
	Selectors []AdminSelector `json:"selectors,omitempty"`
}
