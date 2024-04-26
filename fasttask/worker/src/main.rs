mod bridge;
mod pb;
mod task;
mod executor;

use async_channel::{self, Sender};
use clap::{Args, Parser, Subcommand};
use tokio;
use tokio::runtime::Builder;
use tracing::{self, error};
use tracing_subscriber::{self, EnvFilter};

const FAILED: i32 = 7; // 7=retryable 8=permanent
const QUEUED: i32 = 3;
const RUNNING: i32 = 5;
const SUCCEEDED: i32 = 6;

pub struct TaskContext {
    kill_tx: Sender<()>,
    last_ack_timestamp: u64,
}

#[derive(Debug, Parser)]
#[command(name = "worker", about = "fasttask worker", long_about = None)]
struct Cli {
    #[arg(short, long, value_name = "EXECUTOR_REGISTRATION_ADDR", default_value = "127.0.0.1:15606", help = "endpoint for executor registration service")]
    executor_registration_addr: String,

    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Debug, Subcommand)]
enum Commands {
    Bridge(BridgeArgs),
    Executor(ExecutorArgs),
}

#[derive(Debug, Args)]
pub struct BridgeArgs {
    #[arg(short, long, value_name = "FASTTASK_URL", default_value = "http://localhost:15605", help = "endpoint url for fasttask service")]
    fasttask_url: String,
    #[arg(short, long, value_name = "QUEUE_ID", default_value = "foo", help = "fasttask queue to listen on for tasks")]
    queue_id: String,
    #[arg(short = 'i', long, value_name = "HEARTBEAT_INTERVAL_SECONDS", default_value = "10", help = "interval in seconds to send heartbeat to fasttask service")]
    heartbeat_interval_seconds: u64,
    #[arg(short, long, value_name = "TASK_STATUS_REPORT_INTERVAL_SECONDS", default_value = "10", help = "interval in seconds to buffer task status for heartbeat")]
    task_status_report_interval_seconds: u64,
    #[arg(short, long, value_name = "LAST_ACK_GRACE_PERIOD_SECONDS", default_value = "90", help = "grace period in seconds to wait for last ack before killing task")]
    last_ack_grace_period_seconds: u64,
    #[arg(short, long, value_name = "PARALLELISM", default_value = "1", help = "number of tasks to run in parallel")]
    parallelism: usize,
    #[arg(short, long, value_name = "BACKLOG_LENGTH", default_value = "5", help = "number of tasks to buffer before dropping assignments")]
    backlog_length: usize,
    #[arg(short='r', long, value_name = "FAST_REGISTER_DIR_OVERRIDE", default_value = "/root", help = "directory to decompress flyte fast registration files")]
    fast_register_dir_override: String,
}

#[derive(Debug, Args)]
pub struct ExecutorArgs {
    #[arg(short, long, value_name = "ID", help = "unique identifier for this executor instance")]
    id: usize,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Cli::parse();
    let subscriber = tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .finish();

    tracing::subscriber::with_default(subscriber, || {
        match args.command {
            Some(Commands::Bridge(bridge_args)) => {
                let runtime = Builder::new_current_thread()
                    .enable_all()
                    .build()
                    .unwrap();

                if let Err(e) = runtime.block_on(bridge::run(bridge_args, &args.executor_registration_addr)) {
                    error!("failed to execute bridge: '{}'", e);
                }
            },
            Some(Commands::Executor(executor_args)) => {

                let runtime = Builder::new_current_thread()
                    .enable_io()
                    .build()
                    .unwrap();

                if let Err(e) = runtime.block_on(executor::run(executor_args, &args.executor_registration_addr)) {
                    error!("failed to execute executor: '{}'", e);
                }
            },
            None => error!("unreachable"),
        }
    });

    Ok(())
}
