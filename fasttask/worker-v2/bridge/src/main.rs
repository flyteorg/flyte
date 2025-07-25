use clap::Parser;
use tokio::runtime::Builder;
use tracing::{self, error};
use tracing_subscriber::{self, EnvFilter};
use unionai_actor_bridge::cli::BridgeArgs;
use unionai_actor_bridge::connection::GRPCConnectionBuilder;
use unionai_actor_bridge::heartbeater::PeriodicHeartbeater;
use unionai_actor_bridge::manager::{ExecutionStrategy, MultiProcessManager};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = BridgeArgs::parse();
    let subscriber = tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .finish();

    tracing::subscriber::with_default(subscriber, || {
        let runtime = Builder::new_current_thread().enable_all().build().unwrap();

        let connection = GRPCConnectionBuilder::new(args.fasttask_url.clone());
        let heartbeater = PeriodicHeartbeater::new(
            args.heartbeat_interval_seconds,
            args.queue_id,
            args.worker_id,
        );
        let manager = MultiProcessManager::new(
            args.backlog_length.try_into().unwrap(),
            ExecutionStrategy::TheExecutionStrategy,
            args.executor_registration_addr.clone(),
            args.fast_register_dir_override.clone(),
            args.last_ack_grace_period_seconds,
            args.parallelism.try_into().unwrap(),
            args.task_status_report_interval_seconds,
        );

        if let Err(e) = runtime.block_on(unionai_actor_bridge::bridge::run(
            connection,
            heartbeater,
            manager,
        )) {
            error!("failed to execute bridge: '{}'", e);
        }
    });

    Ok(())
}
