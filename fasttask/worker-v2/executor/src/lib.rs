mod actor_environment;
mod executor;
mod task_args;
mod tester;

use pyo3::exceptions::PyValueError;
use pyo3::prelude::*;
use std::env;
use tracing_subscriber::layer::SubscriberExt;
use tracing_subscriber::util::SubscriberInitExt;
use tracing_subscriber::{self, fmt};
use unionai_actor_bridge::common::Task;

#[pyfunction]
#[pyo3(name = "executor")]
fn executor_py(py: Python) -> PyResult<Bound<PyAny>> {
    let argv = env::args().collect::<Vec<_>>();

    // Handle parsing logic manually to require less dependencies
    if argv.len() < 6 {
        return Err(PyValueError::new_err(
            "unionai-actor-executor --executor-registration-addr ADDR --id ID [--num-workers NUM_WORKERS]",
        ));
    }

    let executor_registration_addr = match argv.get(3) {
        Some(value) => value.clone(),
        None => return Err(PyValueError::new_err("Unable to get address")),
    };
    let id = match argv.get(5) {
        Some(value) => match value.parse::<usize>() {
            Ok(num) => num,
            Err(_) => return Err(PyValueError::new_err("Unable to parse id")),
        },
        None => return Err(PyValueError::new_err("Unable to get id")),
    };
    let num_workers = match argv.get(7) {
        Some(value) => match value.parse::<usize>() {
            Ok(num) => Some(num),
            Err(_) => return Err(PyValueError::new_err("Unable to parse num_workers")),
        },
        None => None, // Default to None if not provided
    };

    tracing_subscriber::registry()
        .with(
            fmt::layer()
                .with_timer(fmt::time::UtcTime::rfc_3339())
                .with_span_events(fmt::format::FmtSpan::FULL)
                .with_ansi(false)
                .with_thread_names(true),
        )
        .init();

    let executor_args = executor::ExecutorArgs {
        executor_registration_addr,
        id,
        num_workers,
    };

    pyo3_async_runtimes::tokio::future_into_py(py, {
        // Create a new Python GIL token inside the async block
        let locals = pyo3_async_runtimes::tokio::get_current_locals(py)?;

        executor::run_worker_pool(executor_args, locals)
    })
}

#[pyfunction]
fn rust_sleep(py: Python) -> PyResult<Bound<PyAny>> {
    println!("Sleeping ...");
    pyo3_async_runtimes::tokio::future_into_py(py, async {
        tokio::time::sleep(std::time::Duration::from_secs(1)).await;
        Ok(())
    })
}

#[pyfunction]
#[pyo3(name = "tester")]
fn tester_py(py: Python) -> PyResult<Bound<PyAny>> {
    let mut argv = env::args();
    let prog = argv.next().unwrap_or_else(|| "tester".to_string());
    let args: Vec<_> = argv.collect();
    if args.is_empty() {
        eprintln!("Usage: {} <command> [args...]", prog);
        std::process::exit(-1);
    }

    // Get environment variables
    // let run_name = env_vars.get("RUN_NAME");
    // let action_name = env_vars.get("ACTION_NAME");
    // let run_base_dir = env_vars.get("_U_RUN_BASE");
    // let org = env_vars.get("_U_ORG_NAME");
    // let project = env_vars.get("FLYTE_INTERNAL_TASK_PROJECT");
    // let domain = env_vars.get("FLYTE_INTERNAL_TASK_DOMAIN");
    let env_vars = std::env::vars()
        .filter(|(k, _)| {
            k.starts_with("_U_")
                || k.starts_with("FLYTE_")
                || k.starts_with("RUN_")
                || k.starts_with("ACTION_")
        })
        .collect::<std::collections::HashMap<_, _>>();

    let task = Task {
        cmd: args,
        additional_distribution: None,
        fast_register_dir: None,
        env_vars: Some(env_vars),
    };

    // Capture env var to run pool or single executor
    let run_pool_v = std::env::var("RUN_POOL").unwrap_or_else(|_| "true".to_string());

    let run_pool = match run_pool_v.as_str() {
        "true" => true,
        "false" => false,
        _ => return Err(PyValueError::new_err("RUN_POOL must be 'true' or 'false'")),
    };
    println!(
        "[Actor Core] Running task with args: {:?} and run_pool: {}",
        task.cmd, run_pool
    );

    pyo3_async_runtimes::tokio::future_into_py(py, {
        // Create a new Python GIL token inside the async block
        let locals = pyo3_async_runtimes::tokio::get_current_locals(py)?;
        tester::run_task(task, run_pool, locals)
    })
}

#[pymodule]
fn _lib(m: &Bound<'_, PyModule>) -> PyResult<()> {
    m.add_function(wrap_pyfunction!(executor_py, m)?)?;
    m.add_function(wrap_pyfunction!(rust_sleep, m)?)?;
    m.add_function(wrap_pyfunction!(tester_py, m)?)?; // Don't forget to register the function
    Ok(())
}
