// @generated
/// Describes a job that can process independent pieces of data concurrently. Multiple copies of the runnable component
/// will be executed concurrently.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ArrayJob {
    /// Defines the maximum number of instances to bring up concurrently at any given point. Note that this is an
    /// optimistic restriction and that, due to network partitioning or other failures, the actual number of currently
    /// running instances might be more. This has to be a positive number if assigned. Default value is size.
    #[prost(int64, tag="1")]
    pub parallelism: i64,
    /// Defines the number of instances to launch at most. This number should match the size of the input if the job
    /// requires processing of all input data. This has to be a positive number.
    /// In the case this is not defined, the back-end will determine the size at run-time by reading the inputs.
    #[prost(int64, tag="2")]
    pub size: i64,
    #[prost(oneof="array_job::SuccessCriteria", tags="3, 4")]
    pub success_criteria: ::core::option::Option<array_job::SuccessCriteria>,
}
/// Nested message and enum types in `ArrayJob`.
pub mod array_job {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum SuccessCriteria {
        /// An absolute number of the minimum number of successful completions of subtasks. As soon as this criteria is met,
        /// the array job will be marked as successful and outputs will be computed. This has to be a non-negative number if
        /// assigned. Default value is size (if specified).
        #[prost(int64, tag="3")]
        MinSuccesses(i64),
        /// If the array job size is not known beforehand, the min_success_ratio can instead be used to determine when an array
        /// job can be marked successful.
        #[prost(float, tag="4")]
        MinSuccessRatio(f32),
    }
}
/// Custom Proto for Dask Plugin.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DaskJob {
    /// Spec for the scheduler pod.
    #[prost(message, optional, tag="1")]
    pub scheduler: ::core::option::Option<DaskScheduler>,
    /// Spec of the default worker group.
    #[prost(message, optional, tag="2")]
    pub workers: ::core::option::Option<DaskWorkerGroup>,
}
/// Specification for the scheduler pod.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DaskScheduler {
    /// Optional image to use. If unset, will use the default image.
    #[prost(string, tag="1")]
    pub image: ::prost::alloc::string::String,
    /// Resources assigned to the scheduler pod.
    #[prost(message, optional, tag="2")]
    pub resources: ::core::option::Option<super::core::Resources>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DaskWorkerGroup {
    /// Number of workers in the group.
    #[prost(uint32, tag="1")]
    pub number_of_workers: u32,
    /// Optional image to use for the pods of the worker group. If unset, will use the default image.
    #[prost(string, tag="2")]
    pub image: ::prost::alloc::string::String,
    /// Resources assigned to the all pods of the worker group.
    /// As per <https://kubernetes.dask.org/en/latest/kubecluster.html?highlight=limit#best-practices> 
    /// it is advised to only set limits. If requests are not explicitly set, the plugin will make
    /// sure to set requests==limits.
    /// The plugin sets ` --memory-limit` as well as `--nthreads` for the workers according to the limit.
    #[prost(message, optional, tag="3")]
    pub resources: ::core::option::Option<super::core::Resources>,
}
/// MPI operator proposal <https://github.com/kubeflow/community/blob/master/proposals/mpi-operator-proposal.md>
/// Custom proto for plugin that enables distributed training using <https://github.com/kubeflow/mpi-operator>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedMpiTrainingTask {
    /// number of worker spawned in the cluster for this job
    #[prost(int32, tag="1")]
    pub num_workers: i32,
    /// number of launcher replicas spawned in the cluster for this job
    /// The launcher pod invokes mpirun and communicates with worker pods through MPI.
    #[prost(int32, tag="2")]
    pub num_launcher_replicas: i32,
    /// number of slots per worker used in hostfile.
    /// The available slots (GPUs) in each pod.
    #[prost(int32, tag="3")]
    pub slots: i32,
}
/// This message works with the 'presto' task type in the SDK and is the object that will be in the 'custom' field
/// of a Presto task's TaskTemplate
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct PrestoQuery {
    #[prost(string, tag="1")]
    pub routing_group: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub catalog: ::prost::alloc::string::String,
    #[prost(string, tag="3")]
    pub schema: ::prost::alloc::string::String,
    #[prost(string, tag="4")]
    pub statement: ::prost::alloc::string::String,
}
/// Custom proto for torch elastic config for distributed training using 
/// <https://github.com/kubeflow/training-operator/blob/master/pkg/apis/kubeflow.org/v1/pytorch_types.go>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ElasticConfig {
    #[prost(string, tag="1")]
    pub rdzv_backend: ::prost::alloc::string::String,
    #[prost(int32, tag="2")]
    pub min_replicas: i32,
    #[prost(int32, tag="3")]
    pub max_replicas: i32,
    #[prost(int32, tag="4")]
    pub nproc_per_node: i32,
    #[prost(int32, tag="5")]
    pub max_restarts: i32,
}
/// Custom proto for plugin that enables distributed training using <https://github.com/kubeflow/pytorch-operator>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedPyTorchTrainingTask {
    /// number of worker replicas spawned in the cluster for this job
    #[prost(int32, tag="1")]
    pub workers: i32,
    /// config for an elastic pytorch job
    /// 
    #[prost(message, optional, tag="2")]
    pub elastic_config: ::core::option::Option<ElasticConfig>,
}
/// Defines a query to execute on a hive cluster.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HiveQuery {
    #[prost(string, tag="1")]
    pub query: ::prost::alloc::string::String,
    #[prost(uint32, tag="2")]
    pub timeout_sec: u32,
    #[prost(uint32, tag="3")]
    pub retry_count: u32,
}
/// Defines a collection of hive queries.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HiveQueryCollection {
    #[prost(message, repeated, tag="2")]
    pub queries: ::prost::alloc::vec::Vec<HiveQuery>,
}
/// This message works with the 'hive' task type in the SDK and is the object that will be in the 'custom' field
/// of a hive task's TaskTemplate
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct QuboleHiveJob {
    #[prost(string, tag="1")]
    pub cluster_label: ::prost::alloc::string::String,
    #[deprecated]
    #[prost(message, optional, tag="2")]
    pub query_collection: ::core::option::Option<HiveQueryCollection>,
    #[prost(string, repeated, tag="3")]
    pub tags: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(message, optional, tag="4")]
    pub query: ::core::option::Option<HiveQuery>,
}
/// RayJobSpec defines the desired state of RayJob
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RayJob {
    /// RayClusterSpec is the cluster template to run the job
    #[prost(message, optional, tag="1")]
    pub ray_cluster: ::core::option::Option<RayCluster>,
    /// runtime_env is base64 encoded.
    /// Ray runtime environments: <https://docs.ray.io/en/latest/ray-core/handling-dependencies.html#runtime-environments>
    #[prost(string, tag="2")]
    pub runtime_env: ::prost::alloc::string::String,
}
/// Define Ray cluster defines the desired state of RayCluster
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RayCluster {
    /// HeadGroupSpecs are the spec for the head pod
    #[prost(message, optional, tag="1")]
    pub head_group_spec: ::core::option::Option<HeadGroupSpec>,
    /// WorkerGroupSpecs are the specs for the worker pods
    #[prost(message, repeated, tag="2")]
    pub worker_group_spec: ::prost::alloc::vec::Vec<WorkerGroupSpec>,
}
/// HeadGroupSpec are the spec for the head pod
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HeadGroupSpec {
    /// Optional. RayStartParams are the params of the start command: address, object-store-memory.
    /// Refer to <https://docs.ray.io/en/latest/ray-core/package-ref.html#ray-start>
    #[prost(map="string, string", tag="1")]
    pub ray_start_params: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
/// WorkerGroupSpec are the specs for the worker pods
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WorkerGroupSpec {
    /// Required. RayCluster can have multiple worker groups, and it distinguishes them by name
    #[prost(string, tag="1")]
    pub group_name: ::prost::alloc::string::String,
    /// Required. Desired replicas of the worker group. Defaults to 1.
    #[prost(int32, tag="2")]
    pub replicas: i32,
    /// Optional. Min replicas of the worker group. MinReplicas defaults to 1.
    #[prost(int32, tag="3")]
    pub min_replicas: i32,
    /// Optional. Max replicas of the worker group. MaxReplicas defaults to maxInt32
    #[prost(int32, tag="4")]
    pub max_replicas: i32,
    /// Optional. RayStartParams are the params of the start command: address, object-store-memory.
    /// Refer to <https://docs.ray.io/en/latest/ray-core/package-ref.html#ray-start>
    #[prost(map="string, string", tag="5")]
    pub ray_start_params: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SparkApplication {
}
/// Nested message and enum types in `SparkApplication`.
pub mod spark_application {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Type {
        Python = 0,
        Java = 1,
        Scala = 2,
        R = 3,
    }
    impl Type {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Type::Python => "PYTHON",
                Type::Java => "JAVA",
                Type::Scala => "SCALA",
                Type::R => "R",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "PYTHON" => Some(Self::Python),
                "JAVA" => Some(Self::Java),
                "SCALA" => Some(Self::Scala),
                "R" => Some(Self::R),
                _ => None,
            }
        }
    }
}
/// Custom Proto for Spark Plugin.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SparkJob {
    #[prost(enumeration="spark_application::Type", tag="1")]
    pub application_type: i32,
    #[prost(string, tag="2")]
    pub main_application_file: ::prost::alloc::string::String,
    #[prost(string, tag="3")]
    pub main_class: ::prost::alloc::string::String,
    #[prost(map="string, string", tag="4")]
    pub spark_conf: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    #[prost(map="string, string", tag="5")]
    pub hadoop_conf: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    /// Executor path for Python jobs.
    #[prost(string, tag="6")]
    pub executor_path: ::prost::alloc::string::String,
    /// Databricks job configuration.
    /// Config structure can be found here. <https://docs.databricks.com/dev-tools/api/2.0/jobs.html#request-structure.>
    #[prost(message, optional, tag="7")]
    pub databricks_conf: ::core::option::Option<::prost_types::Struct>,
    /// Databricks access token. <https://docs.databricks.com/dev-tools/api/latest/authentication.html>
    /// This token can be set in either flytepropeller or flytekit.
    #[prost(string, tag="8")]
    pub databricks_token: ::prost::alloc::string::String,
    /// Domain name of your deployment. Use the form <account>.cloud.databricks.com.
    /// This instance name can be set in either flytepropeller or flytekit.
    #[prost(string, tag="9")]
    pub databricks_instance: ::prost::alloc::string::String,
}
/// Custom proto for plugin that enables distributed training using <https://github.com/kubeflow/tf-operator>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedTensorflowTrainingTask {
    /// number of worker, ps, chief replicas spawned in the cluster for this job
    #[prost(int32, tag="1")]
    pub workers: i32,
    /// PS -> Parameter server
    #[prost(int32, tag="2")]
    pub ps_replicas: i32,
    #[prost(int32, tag="3")]
    pub chief_replicas: i32,
}
/// Represents an Execution that was launched and could be waited on.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Waitable {
    #[prost(message, optional, tag="1")]
    pub wf_exec_id: ::core::option::Option<super::core::WorkflowExecutionIdentifier>,
    #[prost(enumeration="super::core::workflow_execution::Phase", tag="2")]
    pub phase: i32,
    #[prost(string, tag="3")]
    pub workflow_id: ::prost::alloc::string::String,
}
// @@protoc_insertion_point(module)
