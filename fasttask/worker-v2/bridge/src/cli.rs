use clap::Parser;

#[derive(Parser, Debug)]
#[command(name = "worker", about = "fasttask worker", long_about = None)]
pub struct BridgeArgs {
    #[arg(
        short,
        long,
        value_name = "EXECUTOR_REGISTRATION_ADDR",
        default_value = "127.0.0.1:15606",
        help = "endpoint for executor registration service"
    )]
    pub executor_registration_addr: String,
    #[arg(
        short,
        long,
        value_name = "FASTTASK_URL",
        default_value = "http://localhost:15605",
        help = "endpoint url for fasttask service"
    )]
    pub fasttask_url: String,
    #[arg(
        short,
        long,
        value_name = "QUEUE_ID",
        required = true,
        help = "fasttask queue to listen on for tasks"
    )]
    pub queue_id: String,
    #[arg(
        short,
        long,
        value_name = "WORKER_ID",
        required = true,
        help = "worker id to register with fasttask service"
    )]
    pub worker_id: String,
    #[arg(
        short = 'i',
        long,
        value_name = "HEARTBEAT_INTERVAL_SECONDS",
        default_value = "10",
        help = "interval in seconds to send heartbeat to fasttask service"
    )]
    pub heartbeat_interval_seconds: u64,
    #[arg(
        short,
        long,
        value_name = "TASK_STATUS_REPORT_INTERVAL_SECONDS",
        default_value = "10",
        help = "interval in seconds to buffer task status for heartbeat"
    )]
    pub task_status_report_interval_seconds: u64,
    #[arg(
        short,
        long,
        value_name = "LAST_ACK_GRACE_PERIOD_SECONDS",
        default_value = "90",
        help = "grace period in seconds to wait for last ack before killing task"
    )]
    pub last_ack_grace_period_seconds: u64,
    #[arg(
        short,
        long,
        value_name = "PARALLELISM",
        default_value = "1",
        help = "number of tasks to run in parallel"
    )]
    pub parallelism: usize,
    #[arg(
        short,
        long,
        value_name = "BACKLOG_LENGTH",
        default_value = "0",
        help = "suggested number of tasks to buffer for future execution, the actual number may be higher"
    )]
    pub backlog_length: usize,
    #[arg(
        short = 'r',
        long,
        value_name = "FAST_REGISTER_DIR_OVERRIDE",
        default_value = "/root",
        help = "directory to decompress flyte fast registration files"
    )]
    pub fast_register_dir_override: String,
}
