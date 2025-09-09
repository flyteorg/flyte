use anyhow::Result;
use futures::StreamExt;
use tokio::net::TcpListener;
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::{info, warn};

use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize logging
    let stdout_layer = tracing_subscriber::fmt::layer()
        .with_ansi(true) // Enable colors
        .with_target(true) // Omit target/module info for cleaner output
        .with_level(true) // Show log level
        .with_line_number(true); // Omit line numbers
    tracing_subscriber::registry()
        .with(stdout_layer)
        .init();

    let addr = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "127.0.0.1:15606".to_string());

    info!("Starting test server on {}", addr);
    
    let listener = TcpListener::bind(&addr).await?;
    info!("Test server listening on {}", addr);

    loop {
        // Accept a single connection (like the bridge does)
        let (stream, client_addr) = listener.accept().await?;
        info!("Executor connected from {}", client_addr);

        // Set up the same framing that the bridge uses
        let framed = Framed::new(stream, LengthDelimitedCodec::new());
        let (mut _writer, mut reader) = framed.split();

        info!("Connection established, listening for messages...");

        // Just read messages and log them for now
        while let Some(message) = reader.next().await {
            match message {
                Ok(bytes) => {
                    info!(
                    "Received {} bytes from executor: {:?}",
                    bytes.len(),
                    bytes
                );

                    // Try to deserialize as Response to see what the executor is sending
                    if let Ok(response) = bincode::deserialize::<unionai_actor_bridge::common::Response>(&bytes) {
                        info!("Parsed as Response: {:?}", response);
                    } else {
                        warn!("Could not parse as Response, raw bytes: {:?}", bytes);
                    }
                }
                Err(e) => {
                    warn!("Error reading message: {}", e);
                    break;
                }
            }
        }
        info!("Connection closed, test server connection will restart, ctrl+c to terminate");
    }
    Ok(())
}