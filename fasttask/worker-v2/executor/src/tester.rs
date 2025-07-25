use futures::{SinkExt, StreamExt};
use pyo3::exceptions::PyRuntimeError;
use pyo3::PyErr;
use pyo3_async_runtimes::TaskLocals;
use std::println;
use tokio::net::{TcpListener, TcpStream};
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::field::debug;
use tracing::{debug, error};
use tracing_subscriber::util::SubscriberInitExt;
use tracing_subscriber::{fmt, layer::SubscriberExt};
use unionai_actor_bridge::common::Task;

pub async fn run_server_test(addr: String, task: Task) -> Result<(), PyErr> {
    let listener = TcpListener::bind(&addr).await.map_err(|e| {
        PyRuntimeError::new_err(format!("Failed to bind to address {}: {}", addr, e))
    })?;
    println!("Server listening on {}", &addr);

    while let Ok((stream, addr)) = listener.accept().await {
        println!("Client connected from: {}", addr);
        tokio::spawn(handle_client(stream, task.clone()));
    }

    Ok(())
}

async fn handle_client(stream: TcpStream, task: Task) {
    let mut framed = Framed::new(stream, LengthDelimitedCodec::new());

    let task_bytes = bincode::serialize(&task).expect("Failed to serialize task");
    if let Err(e) = framed.send(task_bytes.into()).await {
        error!("Failed to send task to client: {}", e);
        return;
    }

    println!("Sent job");

    // Wait for response from worker
    match framed.next().await {
        Some(Ok(response)) => {
            println!(
                "Received response: {:?}",
                String::from_utf8_lossy(&response)
            );
        }
        Some(Err(e)) => {
            println!("Error receiving response: {}", e);
        }
        None => {
            println!("Client disconnected");
        }
    }

    println!("Client session ended");
}

pub async fn run_task(task: Task, run_pool: bool, locals: TaskLocals) -> Result<(), PyErr> {
    tracing_subscriber::registry()
        .with(
            fmt::layer()
                .with_timer(fmt::time::UtcTime::rfc_3339())
                .with_span_events(fmt::format::FmtSpan::FULL)
                .with_ansi(false)
                .with_thread_names(true),
        )
        .init();

    let executor_args = crate::executor::ExecutorArgs {
        executor_registration_addr: String::from("127.0.0.1:8999"),
        id: 1,
        num_workers: Some(1),
    };

    let j = tokio::spawn(run_server_test(
        executor_args.executor_registration_addr.clone(),
        task,
    ));

    tokio::time::sleep(std::time::Duration::from_secs(1)).await;
    if run_pool {
        println!(
            "[Actor Core] Running worker pool with args: {:?}",
            executor_args
        );
        crate::executor::run_worker_pool(executor_args, locals).await?;
    } else {
        println!(
            "[Actor Core] Running executor with args: {:?}",
            executor_args
        );
        crate::executor::run(executor_args, locals).await?;
    }
    // Create a new Python GIL token inside the async block

    j.await
        .map_err(|e| PyRuntimeError::new_err(e.to_string()))?
}
