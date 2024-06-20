mod executor;

use pyo3::exceptions::PyValueError;
use pyo3::prelude::*;
use std::env;
use tokio::runtime::Builder;
use tracing::{self, error};
use tracing_subscriber::{self, EnvFilter};

#[pyfunction]
#[pyo3(name = "executor")]
fn executor_py(py: Python) -> PyResult<()> {
    let argv = env::args().collect::<Vec<_>>();

    // Handle parsing logic manually to require less dependencies
    if argv.len() != 6 {
        return Err(PyValueError::new_err(
            "unionai-actor-executor --executor-registration-addr ADDR --id ID",
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

    let executor_args = executor::ExecutorArgs {
        executor_registration_addr,
        id,
    };

    let subscriber = tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .finish();

    tracing::subscriber::with_default(subscriber, || {
        let runtime = Builder::new_current_thread().enable_io().build().unwrap();

        if let Err(e) = runtime.block_on(executor::run(py, executor_args)) {
            error!("failed to execute executor: '{}'", e);
        }
    });

    Ok(())
}

#[pymodule]
fn _lib(m: &Bound<'_, PyModule>) -> PyResult<()> {
    m.add_function(wrap_pyfunction!(executor_py, m)?)?;
    Ok(())
}
