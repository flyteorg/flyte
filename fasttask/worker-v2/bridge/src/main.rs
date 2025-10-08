use clap::Parser;
use tokio::signal;
use tokio_util::sync::CancellationToken;
use tracing::{error, info, warn};
use unionai_actor_bridge::cli::BridgeArgs;
use unionai_actor_bridge::connection::GRPCConnectionBuilder;
use unionai_actor_bridge::manager::V2TaskManager;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = BridgeArgs::parse();
    let (_temp_dir, _file_guard, _stdout_guard) =
        unionai_actor_bridge::init_tracing_with_prefix("bridge")?;

    info!("Starting fasttask worker-v2 bridge");
    // Initialize GRPC connection
    let connection = GRPCConnectionBuilder::new(args.fasttask_url.clone());

    // Initialize V2TaskManager
    let mut manager = V2TaskManager::new(
        args.parallelism.try_into().unwrap(),
        args.executor_registration_addr.clone().as_str(),
        args.worker_id.clone(),
        args.queue_id.clone(),
        args.heartbeat_interval_seconds,
        // todo: args.last_ack_grace_period_seconds, // this is the reverse heartbeating
    );

    // Start the executor process and TCP listener
    manager.start().await?;

    // Create cancellation token for graceful shutdown
    let cancellation_token = CancellationToken::new();

    // Setup signal handlers
    let mut sigterm = signal::unix::signal(signal::unix::SignalKind::terminate())?;
    let mut sigint = signal::unix::signal(signal::unix::SignalKind::interrupt())?;

    tokio::select! {
        res = manager.run(connection, cancellation_token.clone()) => {
            manager.send_final_heartbeat(args.fasttask_url.clone()).await;
            info!("Final heartbeat sent following manager run");
            if let Err(e) = res {
                error!(error = %e, "Bridge run finished with error");
                return Err(e.into());
            } else {
                info!("Bridge run completed normally");
            }
        }

        // Handle SIGTERM (what k8s sends before SIGKILL)
        _ = sigterm.recv() => {
            warn!("Received SIGTERM, initiating graceful shutdown");
            cancellation_token.cancel();
            manager.send_final_heartbeat(args.fasttask_url.clone()).await;
        }

        // Handle SIGINT
        _ = sigint.recv() => {
            warn!("Received SIGINT, initiating graceful shutdown");
            cancellation_token.cancel();
            manager.send_final_heartbeat(args.fasttask_url.clone()).await;
        }
    }
    info!("Bridge run completed, exiting...");

    Ok(())
}
