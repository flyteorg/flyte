pub mod datacatalog {
    include!("datacatalog.rs");
}
pub mod flyteidl {

  pub mod admin {
    include!("flyteidl.admin.rs");
  }
  pub mod cache {
    include!("flyteidl.cacheservice.rs");
  }
  pub mod core {
    include!("flyteidl.core.rs");
  }
  pub mod event {
    include!("flyteidl.event.rs");
  }
  pub mod plugins {
    include!("flyteidl.plugins.rs");
    pub mod kubeflow{
      include!("flyteidl.plugins.kubeflow.rs");
    }
  }
  pub mod service {
    include!("flyteidl.service.rs");
  }
}