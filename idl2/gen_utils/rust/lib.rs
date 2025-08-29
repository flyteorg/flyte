// mod test_foo;

// mod serde;

// mod serde_impl;

use pyo3;
use pyo3::prelude::*;
// use crate::HeartbeatResponse;
// include!("flyteidl.common.rs");
// include!("flyteidl.workflow.rs");
// include!("flyteidl.workflow.tonic.rs");
// inculde!("flyteidl.logs.dataplane.rs");
// include!("flyteidl.core.rs");
// include!("google.rpc.rs");
// include!("validate.rs");
use pyo3::pymodule;

// use crate::*;
// Re-export all generated protobuf modules
pub mod flyteidl {

    pub mod common {
        include!("flyteidl.common.rs");
        //         pub mod serde {
        //             include!("flyteidl.common.serde.rs");
        //         }
    }
    pub mod workflow {
        include!("flyteidl.workflow.rs");
        //         pub mod serde {
        //             include!("flyteidl.workflow.serde.rs");
        //         }
        //         pub mod tonic {
        //             include!("flyteidl.workflow.tonic.rs");
        //         }
    }

    pub mod logs {
        pub mod dataplane {
            include!("flyteidl.logs.dataplane.rs");
            //             pub mod serde {
            //                 include!("flyteidl.logs.dataplane.serde.rs");
            //             }
        }
    }

    pub mod core {
        include!("flyteidl.core.rs");
        //         pub mod serde {
        //             include!("flyteidl.core.serde.rs");
        //         }
    }
}

// use pyo3_prost::pyclass_for_prost_struct;
pub mod google {
    pub mod rpc {
        include!("google.rpc.rs");
        //         pub mod serde {
        //             include!("google.rpc.serde.rs");
        //         }
    }
    pub mod protobuf {
        include!(concat!(env!("OUT_DIR"), "/google.protobuf.rs"));
        include!(concat!(env!("OUT_DIR"), "/google.protobuf.serde.rs"));
    }
}

pub mod validate {
    include!("validate.rs");
    //     pub mod serde {
    //         include!("validate.serde.rs");
    //     }
}

// Include the generated Box<T> implementations
include!(concat!(env!("OUT_DIR"), "/boxed_impls.rs"));
// pub mod serde {
//     include!(concat!(env!("OUT_DIR"), "/serde_impls.rs"));
// }
pub mod pymodules {
    include!(concat!(env!("OUT_DIR"), "/pymodules.rs"));
}

include!(concat!(env!("OUT_DIR"), "/serde_impls.rs"));
// include!("serde_impl.rs");

//
// // Re-export commonly used types at the root level for convenience
// pub use flyteidl::common::*;
// pub use flyteidl::workflow::*;
// pub use flyteidl::core::*;
// pub use google::rpc::*;
// pub use validate::*;
// pub use crate::*;

// #[pymodule]
// fn pb_rust(_py: Python, m: &PyModule) -> PyResult<()> {
//     m.add_submodule(flyteidl::workflow::make_module(_py)?)?;
//     // ... all submodules ...
//     Ok(())
// }
