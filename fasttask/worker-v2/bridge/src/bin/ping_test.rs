use anyhow::Result;
use clap::Parser;
use futures::{SinkExt, StreamExt};
use std::collections::HashMap;
use tokio::net::{TcpListener, TcpStream};
use tokio::time::{sleep, Duration, Instant};
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::{error, info, warn};
use unionai_actor_bridge::common::{Response, Task};
use uuid::Uuid;

use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Parser)]
#[command(name = "ping_test")]
#[command(about = "Generate canonical task tests for the executor using real task commands")]
struct Cli {
    /// Address to bind to (where executor will connect)
    #[arg(short, long, default_value = "127.0.0.1:15606")]
    addr: String,
    /// Seconds to wait before sending task
    #[arg(long, default_value = "1")]
    task_wait: u64,
    /// Whether to test cancellation after sending the task
    #[arg(long)]
    test_cancel: bool,
    /// Seconds to wait before cancellation (if test_cancel is true)
    #[arg(long, default_value = "3")]
    cancel_wait: u64,
}

fn create_real_task(task_id: &str) -> Task {
    let mut env_vars = HashMap::new();

    // Environment variables from actual logs
    env_vars.insert("LOG_LEVEL".to_string(), "30".to_string());
    env_vars.insert("ACTION_NAME".to_string(), task_id.to_string());
    env_vars.insert(
        "FLYTE_INTERNAL_NAME".to_string(),
        "say_hello_nested".to_string(),
    );
    env_vars.insert(
        "FLYTE_INTERNAL_DOMAIN".to_string(),
        "development".to_string(),
    );
    env_vars.insert(
        "FLYTE_INTERNAL_TASK_NAME".to_string(),
        "say_hello_nested".to_string(),
    );
    env_vars.insert(
        "FLYTE_INTERNAL_TASK_VERSION".to_string(),
        "2f17eb3023f22fac4058f736aada9bbf".to_string(),
    );
    env_vars.insert(
        "FLYTE_INTERNAL_EXECUTION_DOMAIN".to_string(),
        "development".to_string(),
    );
    env_vars.insert(
        "FLYTE_INTERNAL_TASK_PROJECT".to_string(),
        "testproject".to_string(),
    );
    env_vars.insert(
        "FLYTE_INTERNAL_PROJECT".to_string(),
        "testproject".to_string(),
    );
    env_vars.insert(
        "FLYTE_INTERNAL_EXECUTION_PROJECT".to_string(),
        "testproject".to_string(),
    );
    env_vars.insert("RUN_NAME".to_string(), "test-run-canonical".to_string());
    env_vars.insert("_F_PN".to_string(), "".to_string());
    env_vars.insert("FLYTE_ATTEMPT_NUMBER".to_string(), "0".to_string());
    env_vars.insert(
        "_U_RUN_BASE".to_string(),
        format!("/tmp/test-run-base/{}", task_id),
    );
    env_vars.insert("_U_ORG_NAME".to_string(), "testorg".to_string());
    env_vars.insert(
        "FLYTE_INTERNAL_EXECUTION_ID".to_string(),
        "test-canonical-execution".to_string(),
    );
    env_vars.insert(
        "FLYTE_INTERNAL_VERSION".to_string(),
        "2f17eb3023f22fac4058f736aada9bbf".to_string(),
    );
    env_vars.insert(
        "FLYTE_INTERNAL_TASK_DOMAIN".to_string(),
        "development".to_string(),
    );

    // Command targeting examples/basics/devbox_one.py say_hello_nested
    let cmd = vec![
        "a0".to_string(),
        "--inputs".to_string(),
        // format!("/tmp/test-inputs/{}/inputs.pb", task_id),
        format!("/Users/ytong/actor_debugging/inputs.pb"),
        "--outputs-path".to_string(),
        format!("/tmp/test-outputs/{}", task_id),
        "--version".to_string(),
        "tst-version-123".to_string(),
        "--raw-data-path".to_string(),
        format!("/tmp/test-data/{}/kv/{}", task_id, task_id),
        "--checkpoint-path".to_string(),
        format!("/tmp/test-data/{}/kv/{}/_flytecheckpoints", task_id, task_id),
        "--prev-checkpoint".to_string(),
        "\"\"".to_string(),
        "--run-name".to_string(),
        "test-run-canonical".to_string(),
        "--name".to_string(),
        task_id.to_string(),
        "--image-cache".to_string(),
        "H4sIAAAAAAAC/03MQQrCMBCF4bvMWtLYSRByGZlMpjGYMlJTQUrurq0bV+/fvG+DMlOWa1W9rw8IG9DadF80ZwsB8o0XU3SY6ruJLvkXIaJznqxDZI6UJrHxwkIOfSLxPEI/AUVO8vqzvtLxNrtT5jyGSk2eDXrvH4y8QkGJAAAA".to_string(),
        "--tgz".to_string(),
        "/tmp/test-tgz/canonical-test.tar.gz".to_string(),
        "--dest".to_string(),
        ".".to_string(),
        "--resolver".to_string(),
        "flyte._internal.resolvers.default.DefaultTaskResolver".to_string(),
        "mod".to_string(),
        "basics.devbox_one".to_string(),
        "instance".to_string(),
        "say_hello_nested".to_string(),
    ];

    Task {
        cmd,
        additional_distribution: None,
        fast_register_dir: None,
        env_vars: Some(env_vars),
        unique_task_id: task_id.to_string(),
        cancel: false,
    }
}

fn create_cancel_task(task_id: &str) -> Task {
    Task {
        cmd: Vec::new(),
        additional_distribution: None,
        fast_register_dir: None,
        env_vars: Some(HashMap::new()),
        unique_task_id: task_id.to_string(),
        cancel: true,
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize logging
    let stdout_layer = tracing_subscriber::fmt::layer()
        .with_ansi(true)
        .with_target(true)
        .with_level(true)
        .with_line_number(true);
    tracing_subscriber::registry().with(stdout_layer).init();

    let cli = Cli::parse();

    info!("Starting test server listening on {}", cli.addr);
    info!("Will send task: 'say_hello_nested'");
    if cli.test_cancel {
        info!("Will test cancellation after {} seconds", cli.cancel_wait);
    }

    // Connect to the executor

    let listener = TcpListener::bind(&cli.addr).await?;
    info!("Test server listening on {}", cli.addr);

    loop {
        let (stream, client_addr) = listener.accept().await?;
        info!("Connected to executor!!!");

        let framed = Framed::new(stream, LengthDelimitedCodec::new());
        let (mut writer, mut reader) = framed.split();

        // Start listening for responses
        let response_handler = tokio::spawn(async move {
            let mut received_responses = Vec::new();

            while let Some(message) = reader.next().await {
                match message {
                    Ok(bytes) => {
                        info!("Received {} bytes", bytes.len());

                        if let Ok(response) = bincode::deserialize::<Response>(&bytes) {
                            info!(
                                "Parsed response for task '{}': phase={}, reason={:?}, corrupt={}",
                                response.unique_task_id,
                                response.phase,
                                response.reason,
                                response.executor_corrupt
                            );
                            received_responses.push(response);
                        } else {
                            warn!("Could not parse as Response, raw bytes: {:?}", bytes);
                        }
                    }
                    Err(e) => {
                        error!("Error reading response: {}", e);
                        break;
                    }
                }
            }

            received_responses
        });

        // Generate a unique task ID
        let task_id = format!("ping-test-{}", Uuid::new_v4().to_string());
        info!("Generated task ID: {}", task_id);

        // Wait before sending the task
        if cli.task_wait > 0 {
            info!("Waiting {} seconds before sending task...", cli.task_wait);
            sleep(Duration::from_secs(cli.task_wait)).await;
        }

        // Create and send the real task
        let task = create_real_task(&task_id);
        let task_bytes = bincode::serialize(&task)?;
        let send_time = Instant::now();

        info!("Sending task 'say_hello_nested' with ID '{}'", task_id);
        info!("Task command: {:?}", task.cmd);

        writer.send(task_bytes.into()).await?;

        // Test cancellation if requested
        if cli.test_cancel {
            info!(
                "Waiting {} seconds before sending cancellation...",
                cli.cancel_wait
            );
            sleep(Duration::from_secs(cli.cancel_wait)).await;

            let cancel_task = create_cancel_task(&task_id);
            let cancel_bytes = bincode::serialize(&cancel_task)?;

            info!("Sending cancellation for task ID '{}'", task_id);
            writer.send(cancel_bytes.into()).await?;

            // Give some time for final responses
            sleep(Duration::from_secs(3)).await;
        } else {
            // Wait for task completion
            info!("Waiting for task completion...");
            sleep(Duration::from_secs(10)).await;
        }

        // Close the writer to signal completion
        drop(writer);

        // Wait for all responses
        match response_handler.await {
            Ok(responses) => {
                info!("Test completed. Received {} responses:", responses.len());
                for response in responses {
                    info!(
                        "  Task '{}' -> phase={}, reason={:?}",
                        response.unique_task_id, response.phase, response.reason
                    );
                }
            }
            Err(e) => {
                error!("Error waiting for responses: {}", e);
            }
        }
    }

    Ok(())
}
