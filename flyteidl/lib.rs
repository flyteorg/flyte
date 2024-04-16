pub mod flyteidl {

    pub mod admin {
        include!("gen/pb_rust/flyteidl.admin.rs");
    }
    pub mod cache {
        include!("gen/pb_rust/flyteidl.cacheservice.rs");
    }
    pub mod core {
        include!("gen/pb_rust/flyteidl.core.rs");
    }
    pub mod event {
        include!("gen/pb_rust/flyteidl.event.rs");
    }
    pub mod plugins {
        include!("gen/pb_rust/flyteidl.plugins.rs");
        pub mod kubeflow {
            include!("gen/pb_rust/flyteidl.plugins.kubeflow.rs");
        }
    }
    pub mod service {
        include!("gen/pb_rust/flyteidl.service.rs");
    }
}
