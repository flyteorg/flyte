// @generated
/// HyperparameterScalingType defines the way to increase or decrease the value of the hyperparameter
/// For details, refer to: <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-ranges.html>
/// See examples of these scaling type, refer to: <https://aws.amazon.com/blogs/machine-learning/amazon-sagemaker-automatic-model-tuning-now-supports-random-search-and-hyperparameter-scaling/>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HyperparameterScalingType {
}
/// Nested message and enum types in `HyperparameterScalingType`.
pub mod hyperparameter_scaling_type {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Value {
        Auto = 0,
        Linear = 1,
        Logarithmic = 2,
        Reverselogarithmic = 3,
    }
    impl Value {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Value::Auto => "AUTO",
                Value::Linear => "LINEAR",
                Value::Logarithmic => "LOGARITHMIC",
                Value::Reverselogarithmic => "REVERSELOGARITHMIC",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "AUTO" => Some(Self::Auto),
                "LINEAR" => Some(Self::Linear),
                "LOGARITHMIC" => Some(Self::Logarithmic),
                "REVERSELOGARITHMIC" => Some(Self::Reverselogarithmic),
                _ => None,
            }
        }
    }
}
/// ContinuousParameterRange refers to a continuous range of hyperparameter values, allowing
/// users to specify the search space of a floating-point hyperparameter
/// <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-ranges.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContinuousParameterRange {
    #[prost(double, tag="1")]
    pub max_value: f64,
    #[prost(double, tag="2")]
    pub min_value: f64,
    #[prost(enumeration="hyperparameter_scaling_type::Value", tag="3")]
    pub scaling_type: i32,
}
/// IntegerParameterRange refers to a discrete range of hyperparameter values, allowing
/// users to specify the search space of an integer hyperparameter
/// <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-ranges.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct IntegerParameterRange {
    #[prost(int64, tag="1")]
    pub max_value: i64,
    #[prost(int64, tag="2")]
    pub min_value: i64,
    #[prost(enumeration="hyperparameter_scaling_type::Value", tag="3")]
    pub scaling_type: i32,
}
/// ContinuousParameterRange refers to a continuous range of hyperparameter values, allowing
/// users to specify the search space of a floating-point hyperparameter
/// <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-ranges.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CategoricalParameterRange {
    #[prost(string, repeated, tag="1")]
    pub values: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
/// ParameterRangeOneOf describes a single ParameterRange, which is a one-of structure that can be one of
/// the three possible types: ContinuousParameterRange, IntegerParameterRange, and CategoricalParameterRange.
/// This one-of structure in Flyte enables specifying a Parameter in a type-safe manner
/// See: <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-ranges.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ParameterRangeOneOf {
    #[prost(oneof="parameter_range_one_of::ParameterRangeType", tags="1, 2, 3")]
    pub parameter_range_type: ::core::option::Option<parameter_range_one_of::ParameterRangeType>,
}
/// Nested message and enum types in `ParameterRangeOneOf`.
pub mod parameter_range_one_of {
    #[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum ParameterRangeType {
        #[prost(message, tag="1")]
        ContinuousParameterRange(super::ContinuousParameterRange),
        #[prost(message, tag="2")]
        IntegerParameterRange(super::IntegerParameterRange),
        #[prost(message, tag="3")]
        CategoricalParameterRange(super::CategoricalParameterRange),
    }
}
/// ParameterRanges is a map that maps hyperparameter name to the corresponding hyperparameter range
/// <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-ranges.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ParameterRanges {
    #[prost(map="string, message", tag="1")]
    pub parameter_range_map: ::std::collections::HashMap<::prost::alloc::string::String, ParameterRangeOneOf>,
}
/// The input mode that the algorithm supports. When using the File input mode, SageMaker downloads
/// the training data from S3 to the provisioned ML storage Volume, and mounts the directory to docker
/// volume for training container. When using Pipe input mode, Amazon SageMaker streams data directly
/// from S3 to the container.
/// See: <https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_AlgorithmSpecification.html>
/// For the input modes that different SageMaker algorithms support, see:
/// <https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct InputMode {
}
/// Nested message and enum types in `InputMode`.
pub mod input_mode {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Value {
        File = 0,
        Pipe = 1,
    }
    impl Value {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Value::File => "FILE",
                Value::Pipe => "PIPE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "FILE" => Some(Self::File),
                "PIPE" => Some(Self::Pipe),
                _ => None,
            }
        }
    }
}
/// The algorithm name is used for deciding which pre-built image to point to.
/// This is only required for use cases where SageMaker's built-in algorithm mode is used.
/// While we currently only support a subset of the algorithms, more will be added to the list.
/// See: <https://docs.aws.amazon.com/sagemaker/latest/dg/algos.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AlgorithmName {
}
/// Nested message and enum types in `AlgorithmName`.
pub mod algorithm_name {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Value {
        Custom = 0,
        Xgboost = 1,
    }
    impl Value {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Value::Custom => "CUSTOM",
                Value::Xgboost => "XGBOOST",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "CUSTOM" => Some(Self::Custom),
                "XGBOOST" => Some(Self::Xgboost),
                _ => None,
            }
        }
    }
}
/// Specifies the type of file for input data. Different SageMaker built-in algorithms require different file types of input data
/// See <https://docs.aws.amazon.com/sagemaker/latest/dg/cdf-training.html>
/// <https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct InputContentType {
}
/// Nested message and enum types in `InputContentType`.
pub mod input_content_type {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Value {
        TextCsv = 0,
    }
    impl Value {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Value::TextCsv => "TEXT_CSV",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "TEXT_CSV" => Some(Self::TextCsv),
                _ => None,
            }
        }
    }
}
/// Specifies a metric that the training algorithm writes to stderr or stdout.
/// This object is a pass-through.
/// See this for details: <https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_MetricDefinition.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MetricDefinition {
    /// User-defined name of the metric
    #[prost(string, tag="1")]
    pub name: ::prost::alloc::string::String,
    /// SageMaker hyperparameter tuning parses your algorithm’s stdout and stderr streams to find algorithm metrics
    #[prost(string, tag="2")]
    pub regex: ::prost::alloc::string::String,
}
/// Specifies the training algorithm to be used in the training job
/// This object is mostly a pass-through, with a couple of exceptions include: (1) in Flyte, users don't need to specify
/// TrainingImage; either use the built-in algorithm mode by using Flytekit's Simple Training Job and specifying an algorithm
/// name and an algorithm version or (2) when users want to supply custom algorithms they should set algorithm_name field to
/// CUSTOM. In this case, the value of the algorithm_version field has no effect
/// For pass-through use cases: refer to this AWS official document for more details
/// <https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_AlgorithmSpecification.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AlgorithmSpecification {
    /// The input mode can be either PIPE or FILE
    #[prost(enumeration="input_mode::Value", tag="1")]
    pub input_mode: i32,
    /// The algorithm name is used for deciding which pre-built image to point to
    #[prost(enumeration="algorithm_name::Value", tag="2")]
    pub algorithm_name: i32,
    /// The algorithm version field is used for deciding which pre-built image to point to
    /// This is only needed for use cases where SageMaker's built-in algorithm mode is chosen
    #[prost(string, tag="3")]
    pub algorithm_version: ::prost::alloc::string::String,
    /// A list of metric definitions for SageMaker to evaluate/track on the progress of the training job
    /// See this: <https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_AlgorithmSpecification.html>
    /// and this: <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-metrics.html>
    #[prost(message, repeated, tag="4")]
    pub metric_definitions: ::prost::alloc::vec::Vec<MetricDefinition>,
    /// The content type of the input
    /// See <https://docs.aws.amazon.com/sagemaker/latest/dg/cdf-training.html>
    /// <https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html>
    #[prost(enumeration="input_content_type::Value", tag="5")]
    pub input_content_type: i32,
}
/// When enabling distributed training on a training job, the user should use this message to tell Flyte and SageMaker
/// what kind of distributed protocol he/she wants to use to distribute the work.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DistributedProtocol {
}
/// Nested message and enum types in `DistributedProtocol`.
pub mod distributed_protocol {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Value {
        /// Use this value if the user wishes to use framework-native distributed training interfaces.
        /// If this value is used, Flyte won't configure SageMaker to initialize unnecessary components such as
        /// OpenMPI or Parameter Server.
        Unspecified = 0,
        /// Use this value if the user wishes to use MPI as the underlying protocol for her distributed training job
        /// MPI is a framework-agnostic distributed protocol. It has multiple implementations. Currently, we have only
        /// tested the OpenMPI implementation, which is the recommended implementation for Horovod.
        Mpi = 1,
    }
    impl Value {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Value::Unspecified => "UNSPECIFIED",
                Value::Mpi => "MPI",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNSPECIFIED" => Some(Self::Unspecified),
                "MPI" => Some(Self::Mpi),
                _ => None,
            }
        }
    }
}
/// TrainingJobResourceConfig is a pass-through, specifying the instance type to use for the training job, the
/// number of instances to launch, and the size of the ML storage volume the user wants to provision
/// Refer to SageMaker official doc for more details: <https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_CreateTrainingJob.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TrainingJobResourceConfig {
    /// The number of ML compute instances to use. For distributed training, provide a value greater than 1.
    #[prost(int64, tag="1")]
    pub instance_count: i64,
    /// The ML compute instance type
    #[prost(string, tag="2")]
    pub instance_type: ::prost::alloc::string::String,
    /// The size of the ML storage volume that you want to provision.
    #[prost(int64, tag="3")]
    pub volume_size_in_gb: i64,
    /// When users specify an instance_count > 1, Flyte will try to configure SageMaker to enable distributed training.
    /// If the users wish to use framework-agnostic distributed protocol such as MPI or Parameter Server, this
    /// field should be set to the corresponding enum value
    #[prost(enumeration="distributed_protocol::Value", tag="4")]
    pub distributed_protocol: i32,
}
/// The spec of a training job. This is mostly a pass-through object
/// <https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_CreateTrainingJob.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TrainingJob {
    #[prost(message, optional, tag="1")]
    pub algorithm_specification: ::core::option::Option<AlgorithmSpecification>,
    #[prost(message, optional, tag="2")]
    pub training_job_resource_config: ::core::option::Option<TrainingJobResourceConfig>,
}
/// A pass-through for SageMaker's hyperparameter tuning job
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HyperparameterTuningJob {
    /// The underlying training job that the hyperparameter tuning job will launch during the process
    #[prost(message, optional, tag="1")]
    pub training_job: ::core::option::Option<TrainingJob>,
    /// The maximum number of training jobs that an hpo job can launch. For resource limit purpose.
    /// <https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_ResourceLimits.html>
    #[prost(int64, tag="2")]
    pub max_number_of_training_jobs: i64,
    /// The maximum number of concurrent training job that an hpo job can launch
    /// <https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_ResourceLimits.html>
    #[prost(int64, tag="3")]
    pub max_parallel_training_jobs: i64,
}
/// HyperparameterTuningObjectiveType determines the direction of the tuning of the Hyperparameter Tuning Job
/// with respect to the specified metric.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HyperparameterTuningObjectiveType {
}
/// Nested message and enum types in `HyperparameterTuningObjectiveType`.
pub mod hyperparameter_tuning_objective_type {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Value {
        Minimize = 0,
        Maximize = 1,
    }
    impl Value {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Value::Minimize => "MINIMIZE",
                Value::Maximize => "MAXIMIZE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "MINIMIZE" => Some(Self::Minimize),
                "MAXIMIZE" => Some(Self::Maximize),
                _ => None,
            }
        }
    }
}
/// The target metric and the objective of the hyperparameter tuning.
/// <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-define-metrics.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HyperparameterTuningObjective {
    /// HyperparameterTuningObjectiveType determines the direction of the tuning of the Hyperparameter Tuning Job
    /// with respect to the specified metric.
    #[prost(enumeration="hyperparameter_tuning_objective_type::Value", tag="1")]
    pub objective_type: i32,
    /// The target metric name, which is the user-defined name of the metric specified in the
    /// training job's algorithm specification
    #[prost(string, tag="2")]
    pub metric_name: ::prost::alloc::string::String,
}
/// Setting the strategy used when searching in the hyperparameter space
/// Refer this doc for more details:
/// <https://aws.amazon.com/blogs/machine-learning/amazon-sagemaker-automatic-model-tuning-now-supports-random-search-and-hyperparameter-scaling/>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HyperparameterTuningStrategy {
}
/// Nested message and enum types in `HyperparameterTuningStrategy`.
pub mod hyperparameter_tuning_strategy {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Value {
        Bayesian = 0,
        Random = 1,
    }
    impl Value {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Value::Bayesian => "BAYESIAN",
                Value::Random => "RANDOM",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "BAYESIAN" => Some(Self::Bayesian),
                "RANDOM" => Some(Self::Random),
                _ => None,
            }
        }
    }
}
/// When the training jobs launched by the hyperparameter tuning job are not improving significantly,
/// a hyperparameter tuning job can be stopping early.
/// Note that there's only a subset of built-in algorithms that supports early stopping.
/// see: <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-early-stopping.html>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TrainingJobEarlyStoppingType {
}
/// Nested message and enum types in `TrainingJobEarlyStoppingType`.
pub mod training_job_early_stopping_type {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Value {
        Off = 0,
        Auto = 1,
    }
    impl Value {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Value::Off => "OFF",
                Value::Auto => "AUTO",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "OFF" => Some(Self::Off),
                "AUTO" => Some(Self::Auto),
                _ => None,
            }
        }
    }
}
/// The specification of the hyperparameter tuning process
/// <https://docs.aws.amazon.com/sagemaker/latest/dg/automatic-model-tuning-ex-tuning-job.html#automatic-model-tuning-ex-low-tuning-config>
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct HyperparameterTuningJobConfig {
    /// ParameterRanges is a map that maps hyperparameter name to the corresponding hyperparameter range
    #[prost(message, optional, tag="1")]
    pub hyperparameter_ranges: ::core::option::Option<ParameterRanges>,
    /// Setting the strategy used when searching in the hyperparameter space
    #[prost(enumeration="hyperparameter_tuning_strategy::Value", tag="2")]
    pub tuning_strategy: i32,
    /// The target metric and the objective of the hyperparameter tuning.
    #[prost(message, optional, tag="3")]
    pub tuning_objective: ::core::option::Option<HyperparameterTuningObjective>,
    /// When the training jobs launched by the hyperparameter tuning job are not improving significantly,
    /// a hyperparameter tuning job can be stopping early.
    #[prost(enumeration="training_job_early_stopping_type::Value", tag="4")]
    pub training_job_early_stopping_type: i32,
}
// @@protoc_insertion_point(module)
