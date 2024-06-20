use clap::Parser;
use tokio::runtime::Builder;
use tracing::{self, error};
use tracing_subscriber::{self, EnvFilter};
use unionai_actor_bridge::cli::BridgeArgs;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = BridgeArgs::parse();
    let subscriber = tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .finish();

    tracing::subscriber::with_default(subscriber, || {
        let runtime = Builder::new_current_thread().enable_all().build().unwrap();

        if let Err(e) = runtime.block_on(unionai_actor_bridge::run(args)) {
            error!("failed to execute bridge: '{}'", e);
        }
    });

    Ok(())
}
