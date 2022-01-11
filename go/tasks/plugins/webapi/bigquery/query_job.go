package bigquery

import (
	"strconv"

	"github.com/pkg/errors"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	pluginErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"
	pluginUtils "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/api/bigquery/v2"
)

type QueryJobConfig struct {
	Location  string `json:"location"`
	ProjectID string `json:"projectId"`

	// AllowLargeResults: [Optional] If true and query uses legacy SQL
	// dialect, allows the query to produce arbitrarily large result tables
	// at a slight cost in performance. Requires destinationTable to be set.
	// For standard SQL queries, this flag is ignored and large results are
	// always allowed. However, you must still set destinationTable when
	// result size exceeds the allowed maximum response size.
	AllowLargeResults bool `json:"allowLargeResults,omitempty"`

	// Clustering: [Beta] Clustering specification for the destination
	// table. Must be specified with time-based partitioning, data in the
	// table will be first partitioned and subsequently clustered.
	Clustering *bigquery.Clustering `json:"clustering,omitempty"`

	// CreateDisposition: [Optional] Specifies whether the job is allowed to
	// create new tables. The following values are supported:
	// CREATE_IF_NEEDED: If the table does not exist, BigQuery creates the
	// table. CREATE_NEVER: The table must already exist. If it does not, a
	// 'notFound' error is returned in the job result. The default value is
	// CREATE_IF_NEEDED. Creation, truncation and append actions occur as
	// one atomic update upon job completion.
	CreateDisposition string `json:"createDisposition,omitempty"`

	// DefaultDataset: [Optional] Specifies the default dataset to use for
	// unqualified table names in the query. Note that this does not alter
	// behavior of unqualified dataset names.
	DefaultDataset *bigquery.DatasetReference `json:"defaultDataset,omitempty"`

	// DestinationEncryptionConfiguration: Custom encryption configuration
	// (e.g., Cloud KMS keys).
	DestinationEncryptionConfiguration *bigquery.EncryptionConfiguration `json:"destinationEncryptionConfiguration,omitempty"`

	// DestinationTable: [Optional] Describes the table where the query
	// results should be stored. If not present, a new table will be created
	// to store the results. This property must be set for large results
	// that exceed the maximum response size.
	DestinationTable *bigquery.TableReference `json:"destinationTable,omitempty"`

	// FlattenResults: [Optional] If true and query uses legacy SQL dialect,
	// flattens all nested and repeated fields in the query results.
	// allowLargeResults must be true if this is set to false. For standard
	// SQL queries, this flag is ignored and results are never flattened.
	//
	// Default: true
	FlattenResults *bool `json:"flattenResults,omitempty"`

	// MaximumBillingTier: [Optional] Limits the billing tier for this job.
	// Queries that have resource usage beyond this tier will fail (without
	// incurring a charge). If unspecified, this will be set to your project
	// default.
	//
	// Default: 1
	MaximumBillingTier *int64 `json:"maximumBillingTier,omitempty"`

	// MaximumBytesBilled: [Optional] Limits the bytes billed for this job.
	// Queries that will have bytes billed beyond this limit will fail
	// (without incurring a charge). If unspecified, this will be set to
	// your project default.
	MaximumBytesBilled int64 `json:"maximumBytesBilled,omitempty,string"`

	// Priority: [Optional] Specifies a priority for the query. Possible
	// values include INTERACTIVE and BATCH. The default value is
	// INTERACTIVE.
	Priority string `json:"priority,omitempty"`

	// Query: [Required] SQL query text to execute. The useLegacySql field
	// can be used to indicate whether the query uses legacy SQL or standard
	// SQL.
	Query string `json:"query,omitempty"`

	// SchemaUpdateOptions: Allows the schema of the destination table to be
	// updated as a side effect of the query job. Schema update options are
	// supported in two cases: when writeDisposition is WRITE_APPEND; when
	// writeDisposition is WRITE_TRUNCATE and the destination table is a
	// partition of a table, specified by partition decorators. For normal
	// tables, WRITE_TRUNCATE will always overwrite the schema. One or more
	// of the following values are specified: ALLOW_FIELD_ADDITION: allow
	// adding a nullable field to the schema. ALLOW_FIELD_RELAXATION: allow
	// relaxing a required field in the original schema to nullable.
	SchemaUpdateOptions []string `json:"schemaUpdateOptions,omitempty"`

	// TableDefinitions: [Optional] If querying an external data source
	// outside of BigQuery, describes the data format, location and other
	// properties of the data source. By defining these properties, the data
	// source can then be queried as if it were a standard BigQuery table.
	TableDefinitions map[string]bigquery.ExternalDataConfiguration `json:"tableDefinitions,omitempty"`

	// TimePartitioning: Time-based partitioning specification for the
	// destination table. Only one of timePartitioning and rangePartitioning
	// should be specified.
	TimePartitioning *bigquery.TimePartitioning `json:"timePartitioning,omitempty"`

	// UseLegacySQL: Specifies whether to use BigQuery's legacy SQL dialect
	// for this query. The default value is true. If set to false, the query
	// will use BigQuery's standard SQL:
	// https://cloud.google.com/bigquery/sql-reference/ When useLegacySql is
	// set to false, the value of flattenResults is ignored; query will be
	// run as if flattenResults is false.
	//
	// Default: true
	UseLegacySQL *bool `json:"useLegacySql,omitempty"`

	// UseQueryCache: [Optional] Whether to look for the result in the query
	// cache. The query cache is a best-effort cache that will be flushed
	// whenever tables in the query are modified. Moreover, the query cache
	// is only available when a query does not have a destination table
	// specified. The default value is true.
	//
	// Default: true
	UseQueryCache *bool `json:"useQueryCache,omitempty"`

	// UserDefinedFunctionResources: Describes user-defined function
	// resources used in the query.
	UserDefinedFunctionResources []*bigquery.UserDefinedFunctionResource `json:"userDefinedFunctionResources,omitempty"`

	// WriteDisposition: [Optional] Specifies the action that occurs if the
	// destination table already exists. The following values are supported:
	// WRITE_TRUNCATE: If the table already exists, BigQuery overwrites the
	// table data and uses the schema from the query result. WRITE_APPEND:
	// If the table already exists, BigQuery appends the data to the table.
	// WRITE_EMPTY: If the table already exists and contains data, a
	// 'duplicate' error is returned in the job result. The default value is
	// WRITE_EMPTY. Each action is atomic and only occurs if BigQuery is
	// able to complete the job successfully. Creation, truncation and
	// append actions occur as one atomic update upon job completion.
	WriteDisposition string `json:"writeDisposition,omitempty"`
}

func unmarshalQueryJobConfig(structObj *structpb.Struct) (*QueryJobConfig, error) {
	queryJobConfig := QueryJobConfig{}
	err := pluginUtils.UnmarshalStructToObj(structObj, &queryJobConfig)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal QueryJobConfig")
	}

	return &queryJobConfig, nil
}

func getJobConfigurationQuery(custom *QueryJobConfig, inputs *flyteIdlCore.LiteralMap) (*bigquery.JobConfigurationQuery, error) {
	queryParameters, err := getQueryParameters(inputs.Literals)

	if err != nil {
		return nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "unable build query parameters [%v]", err.Error())
	}

	// BigQuery supports query parameters to help prevent SQL injection when queries are constructed using user input.
	// This feature is only available with standard SQL syntax. For more detail: https://cloud.google.com/bigquery/docs/parameterized-queries
	useLegacySQL := false
	return &bigquery.JobConfigurationQuery{
		AllowLargeResults:                  custom.AllowLargeResults,
		Clustering:                         custom.Clustering,
		CreateDisposition:                  custom.CreateDisposition,
		DefaultDataset:                     custom.DefaultDataset,
		DestinationEncryptionConfiguration: custom.DestinationEncryptionConfiguration,
		DestinationTable:                   custom.DestinationTable,
		FlattenResults:                     custom.FlattenResults,
		MaximumBillingTier:                 custom.MaximumBillingTier,
		MaximumBytesBilled:                 custom.MaximumBytesBilled,
		ParameterMode:                      "NAMED",
		Priority:                           custom.Priority,
		Query:                              custom.Query,
		QueryParameters:                    queryParameters,
		SchemaUpdateOptions:                custom.SchemaUpdateOptions,
		TableDefinitions:                   custom.TableDefinitions,
		TimePartitioning:                   custom.TimePartitioning,
		UseLegacySql:                       &useLegacySQL,
		UseQueryCache:                      custom.UseQueryCache,
		UserDefinedFunctionResources:       custom.UserDefinedFunctionResources,
		WriteDisposition:                   custom.WriteDisposition,
	}, nil
}

func getQueryParameters(literalMap map[string]*flyteIdlCore.Literal) ([]*bigquery.QueryParameter, error) {
	queryParameters := make([]*bigquery.QueryParameter, len(literalMap))

	i := 0
	for name, literal := range literalMap {
		parameterType, parameterValue, err := getQueryParameter(literal)

		if err != nil {
			return nil, err
		}

		queryParameters[i] = &bigquery.QueryParameter{
			Name:           name,
			ParameterType:  parameterType,
			ParameterValue: parameterValue,
		}

		i++
	}

	return queryParameters, nil
}

// read more about parameterized queries: https://cloud.google.com/bigquery/docs/parameterized-queries

func getQueryParameter(literal *flyteIdlCore.Literal) (*bigquery.QueryParameterType, *bigquery.QueryParameterValue, error) {
	if scalar := literal.GetScalar(); scalar != nil {
		if primitive := scalar.GetPrimitive(); primitive != nil {
			switch primitive.Value.(type) {
			case *flyteIdlCore.Primitive_Integer:
				integerType := bigquery.QueryParameterType{Type: "INT64"}
				integerValue := bigquery.QueryParameterValue{
					Value: strconv.FormatInt(primitive.GetInteger(), 10),
				}

				return &integerType, &integerValue, nil

			case *flyteIdlCore.Primitive_StringValue:
				stringType := bigquery.QueryParameterType{Type: "STRING"}
				stringValue := bigquery.QueryParameterValue{
					Value: primitive.GetStringValue(),
				}

				return &stringType, &stringValue, nil

			case *flyteIdlCore.Primitive_FloatValue:
				floatType := bigquery.QueryParameterType{Type: "FLOAT64"}
				floatValue := bigquery.QueryParameterValue{
					Value: strconv.FormatFloat(primitive.GetFloatValue(), 'f', -1, 64),
				}

				return &floatType, &floatValue, nil

			case *flyteIdlCore.Primitive_Boolean:
				boolType := bigquery.QueryParameterType{Type: "BOOL"}

				if primitive.GetBoolean() {
					return &boolType, &bigquery.QueryParameterValue{
						Value: "TRUE",
					}, nil
				}

				return &boolType, &bigquery.QueryParameterValue{
					Value: "FALSE",
				}, nil
			}
		}
	}

	return nil, nil, pluginErrors.Errorf(pluginErrors.BadTaskSpecification, "unsupported literal [%v]", literal)
}
