pub mod cli;
pub mod common;
pub mod connection;
pub mod manager;
mod pb;
use std::io;
use tempfile::TempDir;
use tracing_appender::non_blocking::WorkerGuard;
use tracing_subscriber::filter::LevelFilter;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

pub fn init_tracing_with_prefix(
    filename_prefix: &str,
) -> Result<(TempDir, WorkerGuard, WorkerGuard), Box<dyn std::error::Error>> {
    // Create temp directory for logs
    let temp_dir = tempfile::tempdir().expect("Failed to create temp directory for logs");
    let temp_path = temp_dir.path().to_owned();

    let filter = EnvFilter::builder()
        .with_default_directive(LevelFilter::INFO.into())
        .from_env_lossy();

    // Create rotated file appender in temp directory with prefix
    let log_filename = format!("{}.log", filename_prefix);
    let file_appender = tracing_appender::rolling::hourly(&temp_path, &log_filename);
    let (non_blocking_file, file_guard) = tracing_appender::non_blocking(file_appender);

    // Create stdout writer
    let (non_blocking_stdout, stdout_guard) = tracing_appender::non_blocking(io::stdout());

    // File layer - JSON format for structured logging
    let file_layer = tracing_subscriber::fmt::layer()
        .with_timer(tracing_subscriber::fmt::time::UtcTime::rfc_3339())
        .with_span_events(
            tracing_subscriber::fmt::format::FmtSpan::NEW
                | tracing_subscriber::fmt::format::FmtSpan::CLOSE,
        )
        .with_ansi(false)
        .with_thread_names(true)
        .with_thread_ids(true)
        .with_file(true)
        .with_line_number(true)
        .with_target(true)
        .json()
        .with_writer(non_blocking_file);

    // Stdout layer - JSON format for k8s
    let stdout_layer = tracing_subscriber::fmt::layer()
        .with_timer(tracing_subscriber::fmt::time::UtcTime::rfc_3339())
        .with_span_events(
            tracing_subscriber::fmt::format::FmtSpan::NEW
                | tracing_subscriber::fmt::format::FmtSpan::CLOSE,
        )
        .with_ansi(false)
        .with_thread_names(true)
        .with_thread_ids(true)
        .with_file(true)
        .with_line_number(true)
        .with_target(true)
        .json()
        .with_writer(non_blocking_stdout);

    tracing_subscriber::registry()
        .with(filter)
        .with(file_layer)
        .with(stdout_layer)
        .init();

    // Print where logs are being written for debugging
    eprintln!(
        "Logs being written to: {}/{}",
        temp_path.display(),
        log_filename
    );

    Ok((temp_dir, file_guard, stdout_guard))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::Once;
    use tracing::{debug, error, info, trace, warn};

    static INIT: Once = Once::new();

    fn setup_tracing() {
        INIT.call_once(|| {
            let _ = init_tracing_with_prefix("test");
        });
    }

    #[test]
    fn test_init_tracing() {
        setup_tracing();

        println!("Testing tracing output:");

        // Test that we can emit log messages at different levels
        trace!("trace message");
        debug!("debug message");
        info!("info message");
        warn!("warning message");
        error!("error message");

        println!("Tracing test completed - logs should appear above");
    }

    #[tokio::test]
    async fn test_tracing_with_spans() {
        setup_tracing();

        let span = tracing::span!(tracing::Level::INFO, "test_span", test_field = "test_value");
        let _guard = span.enter();

        info!("message within span");

        // Test nested spans
        let nested_span = tracing::span!(tracing::Level::DEBUG, "nested_span");
        let _nested_guard = nested_span.enter();

        debug!("message within nested span");
    }
}
