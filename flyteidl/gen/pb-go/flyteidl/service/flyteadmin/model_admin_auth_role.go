/*
 * flyteidl/service/admin.proto
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: version not set
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package flyteadmin

// Defines permissions associated with executions created by this launch plan spec. Use either of these roles when they have permissions required by your workflow execution. Deprecated.
type AdminAuthRole struct {
	// Defines an optional iam role which will be used for tasks run in executions created with this launch plan.
	AssumableIamRole string `json:"assumable_iam_role,omitempty"`
	// Defines an optional kubernetes service account which will be used for tasks run in executions created with this launch plan.
	KubernetesServiceAccount string `json:"kubernetes_service_account,omitempty"`
}
