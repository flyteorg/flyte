// @generated
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct RunPolicy {
    /// Defines the policy to kill pods after the job completes. Default to None.
    #[prost(enumeration="CleanPodPolicy", tag="1")]
    pub clean_pod_policy: i32,
    /// TTL to clean up jobs. Default to infinite.
    #[prost(int32, tag="2")]
    pub ttl_seconds_after_finished: i32,
    /// Specifies the duration in seconds relative to the startTime that the job may be active
    /// before the system tries to terminate it; value must be positive integer.
    #[prost(int32, tag="3")]
    pub active_deadline_seconds: i32,
    /// Number of retries before marking this job failed.
    #[prost(int32, tag="4")]
    pub backoff_limit: i32,
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum RestartPolicy {
    Never = 0,
    OnFailure = 1,
    Always = 2,
}
impl RestartPolicy {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            RestartPolicy::Never => "RESTART_POLICY_NEVER",
            RestartPolicy::OnFailure => "RESTART_POLICY_ON_FAILURE",
            RestartPolicy::Always => "RESTART_POLICY_ALWAYS",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "RESTART_POLICY_NEVER" => Some(Self::Never),
            "RESTART_POLICY_ON_FAILURE" => Some(Self::OnFailure),
            "RESTART_POLICY_ALWAYS" => Some(Self::Always),
            _ => None,
        }
    }
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum CleanPodPolicy {
    CleanpodPolicyNone = 0,
    CleanpodPolicyRunning = 1,
    CleanpodPolicyAll = 2,
}
impl CleanPodPolicy {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            CleanPodPolicy::CleanpodPolicyNone => "CLEANPOD_POLICY_NONE",
            CleanPodPolicy::CleanpodPolicyRunning => "CLEANPOD_POLICY_RUNNING",
            CleanPodPolicy::CleanpodPolicyAll => "CLEANPOD_POLICY_ALL",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "CLEANPOD_POLICY_NONE" => Some(Self::CleanpodPolicyNone),
            "CLEANPOD_POLICY_RUNNING" => Some(Self::CleanpodPolicyRunning),
            "CLEANPOD_POLICY_ALL" => Some(Self::CleanpodPolicyAll),
            _ => None,
        }
    }
}
/// Proto for plugin that enables distributed training using <https://github.com/kubeflow/mpi-operator>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedMpiTrainingTask {
    /// Worker replicas spec
    #[prost(message, optional, tag="1")]
    pub worker_replicas: ::core::option::Option<DistributedMpiTrainingReplicaSpec>,
    /// Master replicas spec
    #[prost(message, optional, tag="2")]
    pub launcher_replicas: ::core::option::Option<DistributedMpiTrainingReplicaSpec>,
    /// RunPolicy encapsulates various runtime policies of the distributed training
    /// job, for example how to clean up resources and how long the job can stay
    /// active.
    #[prost(message, optional, tag="3")]
    pub run_policy: ::core::option::Option<RunPolicy>,
    /// Number of slots per worker
    #[prost(int32, tag="4")]
    pub slots: i32,
}
/// Replica specification for distributed MPI training
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedMpiTrainingReplicaSpec {
    /// Number of replicas
    #[prost(int32, tag="1")]
    pub replicas: i32,
    /// Image used for the replica group
    #[prost(string, tag="2")]
    pub image: ::prost::alloc::string::String,
    /// Resources required for the replica group
    #[prost(message, optional, tag="3")]
    pub resources: ::core::option::Option<super::super::core::Resources>,
    /// Restart policy determines whether pods will be restarted when they exit
    #[prost(enumeration="RestartPolicy", tag="4")]
    pub restart_policy: i32,
    /// MPI sometimes requires different command set for different replica groups
    #[prost(string, repeated, tag="5")]
    pub command: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
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
/// Proto for plugin that enables distributed training using <https://github.com/kubeflow/pytorch-operator>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedPyTorchTrainingTask {
    /// Worker replicas spec
    #[prost(message, optional, tag="1")]
    pub worker_replicas: ::core::option::Option<DistributedPyTorchTrainingReplicaSpec>,
    /// Master replicas spec, master replicas can only have 1 replica
    #[prost(message, optional, tag="2")]
    pub master_replicas: ::core::option::Option<DistributedPyTorchTrainingReplicaSpec>,
    /// RunPolicy encapsulates various runtime policies of the distributed training
    /// job, for example how to clean up resources and how long the job can stay
    /// active.
    #[prost(message, optional, tag="3")]
    pub run_policy: ::core::option::Option<RunPolicy>,
    /// config for an elastic pytorch job
    #[prost(message, optional, tag="4")]
    pub elastic_config: ::core::option::Option<ElasticConfig>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedPyTorchTrainingReplicaSpec {
    /// Number of replicas
    #[prost(int32, tag="1")]
    pub replicas: i32,
    /// Image used for the replica group
    #[prost(string, tag="2")]
    pub image: ::prost::alloc::string::String,
    /// Resources required for the replica group
    #[prost(message, optional, tag="3")]
    pub resources: ::core::option::Option<super::super::core::Resources>,
    /// RestartPolicy determines whether pods will be restarted when they exit
    #[prost(enumeration="RestartPolicy", tag="4")]
    pub restart_policy: i32,
}
/// Proto for plugin that enables distributed training using <https://github.com/kubeflow/tf-operator>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedTensorflowTrainingTask {
    /// Worker replicas spec
    #[prost(message, optional, tag="1")]
    pub worker_replicas: ::core::option::Option<DistributedTensorflowTrainingReplicaSpec>,
    /// Parameter server replicas spec
    #[prost(message, optional, tag="2")]
    pub ps_replicas: ::core::option::Option<DistributedTensorflowTrainingReplicaSpec>,
    /// Chief replicas spec
    #[prost(message, optional, tag="3")]
    pub chief_replicas: ::core::option::Option<DistributedTensorflowTrainingReplicaSpec>,
    /// RunPolicy encapsulates various runtime policies of the distributed training
    /// job, for example how to clean up resources and how long the job can stay
    /// active.
    #[prost(message, optional, tag="4")]
    pub run_policy: ::core::option::Option<RunPolicy>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedTensorflowTrainingReplicaSpec {
    /// Number of replicas
    #[prost(int32, tag="1")]
    pub replicas: i32,
    /// Image used for the replica group
    #[prost(string, tag="2")]
    pub image: ::prost::alloc::string::String,
    /// Resources required for the replica group
    #[prost(message, optional, tag="3")]
    pub resources: ::core::option::Option<super::super::core::Resources>,
    /// RestartPolicy Determines whether pods will be restarted when they exit
    #[prost(enumeration="RestartPolicy", tag="4")]
    pub restart_policy: i32,
}
// @@protoc_insertion_point(module)
