use crate::task_args::TaskArgs;
use pyo3::prelude::*;
use pyo3::types::PyDict;
use pyo3_async_runtimes::tokio::{future_into_py_with_locals, get_current_locals};
use pyo3_async_runtimes::into_future_with_locals;
use pyo3_async_runtimes::TaskLocals;
use std::collections::HashMap;
use std::time::Duration;
use pyo3::exceptions::PyValueError;
use tokio::sync::Mutex;
use tokio::time::timeout;
use tokio_util::sync::CancellationToken;
use tracing::{debug, info, warn};
use tracing_attributes::instrument;
use unionai_actor_bridge::common::Task;
use crate::TaskError;

// todo: make this 180
const GRACE: Duration = Duration::from_secs(10);

pyo3::create_exception!(async_bridge, CleanupTimeoutError, pyo3::exceptions::PyRuntimeError);


/// Temporary structure to hold the task and its associated code bundle.
/// This is usually passed down to the runtime
pub struct TaskAndBundle {
    pub task: Py<PyAny>,
    pub code_bundle: Py<PyAny>,
}

/// This struct holds resolver information for the task
/// It uses a string for the resolver name and a vector of strings for the arguments
/// This is also used as the lookup key in the task cache, for tasks loaded from code bundle.
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct ResolvedTaskKey {
    pub resolver: String,
    pub resolver_args: Vec<String>,
}

pub struct TGZBundle {
    pub tgz: String,
    pub code_bundle: Py<PyAny>,
}

/// This structure holds the tasks loaded from a TGZ file.
pub struct TGZTasks {
    pub tgz_bundle: TGZBundle,
    pub task_cache: Mutex<HashMap<ResolvedTaskKey, Py<PyAny>>>,
}

impl TGZTasks {
    pub fn new(tgz_bundle: TGZBundle) -> Self {
        Self {
            tgz_bundle,
            task_cache: Mutex::new(HashMap::new()),
        }
    }

    /// Get a task from the cache or load it if not present
    pub async fn get_or_load_task(&self, key: &ResolvedTaskKey) -> Result<TaskAndBundle, PyErr> {
        let mut cache = self.task_cache.lock().await;
        if let Some(task) = cache.get(key) {
            return Ok(Python::with_gil(|py| TaskAndBundle {
                task: task.clone_ref(py),
                code_bundle: self.tgz_bundle.code_bundle.clone_ref(py),
            }));
        }

        Python::with_gil(|py| {
            let rusty = py.import("flyte._internal.runtime.rusty")?;
            let result = rusty.call_method1(
                "load_task_from_code_bundle",
                (
                    key.resolver.clone().into_pyobject(py).unwrap(),
                    key.resolver_args.clone().into_pyobject(py).unwrap(),
                ),
            )?;
            let task: Py<PyAny> = result.extract().unwrap();
            cache.insert(key.clone(), task.clone_ref(py));
            Ok(TaskAndBundle {
                task,
                code_bundle: self.tgz_bundle.code_bundle.clone_ref(py),
            })
        })
    }
}

/// This struct holds a task loaded from a PKL file
pub struct PKLTask {
    pub pkl: String,
    pub code_bundle: Py<PyAny>,
    pub task: Py<PyAny>,
}

/// This struct holds a cache of tasks loaded from PKL files. Ideally the cache should be an
/// LRU Cache.
pub struct PKLTasks {
    pub task_cache: Mutex<HashMap<String, PKLTask>>,
}

impl PKLTasks {
    pub fn new() -> Self {
        Self {
            task_cache: Mutex::new(HashMap::new()),
        }
    }

    /// Get a task from the cache or load it if not present
    pub async fn get_or_load_task(
        &self,
        destination: &str,
        version: &str,
        pkl: &str,
        locals: &TaskLocals,
    ) -> Result<TaskAndBundle, PyErr> {
        let mut cache = self.task_cache.lock().await;
        if let Some(pkl_task) = cache.get(pkl) {
            return Ok(Python::with_gil(|py| TaskAndBundle {
                task: pkl_task.task.clone_ref(py),
                code_bundle: pkl_task.code_bundle.clone_ref(py),
            }));
        }

        let fut = Python::with_gil(|py| {
            let rusty = py.import("flyte._internal.runtime.rusty")?;
            pyo3_async_runtimes::into_future_with_locals(
                locals,
                rusty.call_method1(
                    "download_load_pkl",
                    (
                        destination.into_pyobject(py).unwrap(),
                        version.into_pyobject(py).unwrap(),
                        pkl.into_pyobject(py).unwrap(),
                    ),
                )?,
            )
        })?;

        let v = fut.await?;

        Python::with_gil(|py| {
            let (code_bundle, task): (Py<PyAny>, Py<PyAny>) = v.extract(py)?;
            cache.insert(
                pkl.parse().unwrap(),
                PKLTask {
                    pkl: pkl.to_string(),
                    code_bundle: code_bundle.clone_ref(py),
                    task: task.clone_ref(py),
                },
            );
            Ok(TaskAndBundle { task, code_bundle })
        })
    }
}

/// An ActorEnvironment is a container for all the loaded tasks and the controller
pub struct ActorEnvironment {
    /// Task Templates loaded from a bundled code source
    tgz_tasks: Mutex<Option<TGZTasks>>,
    /// Task Templates loaded from a PKL file
    pkl_tasks: PKLTasks,
    /// Controller for executing tasks
    controller: Py<PyAny>,
    /// Task locals for the actor environment
    locals: TaskLocals,
}

impl ActorEnvironment {
    /// Create a new actor environment with the specified controller
    pub fn new(controller: Py<PyAny>, locals: TaskLocals) -> Self {
        Self {
            tgz_tasks: Mutex::new(None),
            pkl_tasks: PKLTasks::new(),
            controller,
            locals,
        }
    }

    /// Download and load the task
    #[instrument(skip(self), fields(tgz))]
    pub async fn download_tgz(
        &self,
        dest: &str,
        version: &str,
        tgz: &str,
    ) -> Result<Py<PyAny>, PyErr> {
        debug!("Downloading code bundle");
        let fut = Python::with_gil(|py| {
            let rusty = py.import("flyte._internal.runtime.rusty")?;
            pyo3_async_runtimes::into_future_with_locals(
                &self.locals,
                rusty.call_method1(
                    "download_tgz",
                    (
                        dest.into_pyobject(py).unwrap(),
                        version.into_pyobject(py).unwrap(),
                        tgz.into_pyobject(py).unwrap(),
                    ),
                )?,
            )
        })?;
        let v = fut.await?;

        // Store the results
        Ok(Python::with_gil(|py| {
            let code_bundle: Py<PyAny> = v.extract(py).unwrap();
            debug!("Code bundle downloaded successfully");
            code_bundle.clone_ref(py)
        }))
    }

    // Move instrument to past the guard.
    #[instrument(skip(self, task), fields(tgz))]
    async fn get_task_and_bundle_for_tgz(
        &self,
        tgz: &str,
        task: &TaskArgs,
    ) -> Result<TaskAndBundle, PyErr> {
        if task.resolver.is_none() && task.resolver_args.is_empty() {
            return Err(PyErr::new::<pyo3::exceptions::PyValueError, _>(
                "Task with TGZ must have resolver and resolver_args specified",
            ));
        }

        let mut guard = self.tgz_tasks.lock().await;
        if guard.is_none() {
            // Download the TGZ tasks if not already set
            debug!("TGZ tasks not set, downloading");
            let code_bundle = self
                .download_tgz(
                    task.dest.clone().unwrap().as_str(),
                    task.version.clone().unwrap().as_str(),
                    tgz,
                )
                .await?;
            let tgz_tasks = TGZTasks::new(TGZBundle {
                tgz: tgz.to_string(),
                code_bundle,
            });
            guard.replace(tgz_tasks);
            debug!("TGZ tasks initialized");
        }

        if let Some(tgz_task) = guard.as_ref() {
            // check if tgz_tasks if the task has matching tgz value, if not return error
            if tgz != tgz_task.tgz_bundle.tgz.as_str() {
                return Err(PyErr::new::<pyo3::exceptions::PyValueError, _>(
                    "Task TGZ does not match the configured TGZ tasks",
                ));
            }

            let key = ResolvedTaskKey {
                resolver: task.resolver.clone().unwrap_or_default(),
                resolver_args: task.resolver_args.clone(),
            };
            return tgz_task.get_or_load_task(&key).await;
        }

        Err(PyErr::new::<pyo3::exceptions::PyValueError, _>(
            "No TGZ tasks available or task not found",
        ))
    }

    async fn load_task(&self, task: &TaskArgs) -> Result<TaskAndBundle, PyErr> {
        if task.tgz.is_none() && task.pkl.is_none() {
            return Err(PyErr::new::<pyo3::exceptions::PyValueError, _>(
                "Task must have either TGZ or PKL specified",
            ));
        }

        if let Some(tgz) = &task.tgz {
            return self.get_task_and_bundle_for_tgz(tgz, task).await;
        }

        if let Some(pkl) = &task.pkl {
            return self
                .pkl_tasks
                .get_or_load_task(
                    task.dest.clone().unwrap().as_str(),
                    task.version.clone().unwrap().as_str(),
                    pkl,
                    &self.locals,
                )
                .await;
        }

        Err(PyErr::new::<pyo3::exceptions::PyValueError, _>(
            "No valid task source found",
        ))
    }

    #[instrument(skip(self, task), fields(task.id = %task.unique_task_id), level = "info")]
    pub async fn run(&self, task: Task, token: CancellationToken) -> Result<Py<PyAny>, TaskError> {
        let args = TaskArgs::from_command(&task.cmd)
            .map_err(|e| TaskError::Python(PyErr::new::<pyo3::exceptions::PyValueError, _>(e)))?;

        let task_and_bundle = self.load_task(&args).await.map_err(|e| TaskError::Python(e))?;

        let env_vars = task.env_vars.unwrap_or_default();
        let run_name = env_vars.get("RUN_NAME");
        let action_name = env_vars.get("ACTION_NAME");
        let run_base_dir = env_vars.get("_U_RUN_BASE");
        let org = env_vars.get("_U_ORG_NAME");
        let project = env_vars.get("FLYTE_INTERNAL_TASK_PROJECT");
        let domain = env_vars.get("FLYTE_INTERNAL_TASK_DOMAIN");

        debug!(
            run_name = ?run_name,
            action_name = ?action_name,
            org = ?org,
            project = ?project,
            domain = ?domain,
            "Running task with environment context"
        );

        let (py_task, fut) = Python::with_gil(|py| {
            let rusty = py.import("flyte._internal.runtime.rusty")?;

            let kwargs = PyDict::new(py);
            kwargs.set_item("task", task_and_bundle.task.as_ref())?;
            kwargs.set_item("controller", self.controller.as_ref())?;
            kwargs.set_item("code_bundle", task_and_bundle.code_bundle.as_ref())?;
            // Set all the arguments env variables into the kwargs dictionary
            kwargs.set_item("org", org.clone().into_pyobject(py).unwrap())?;
            kwargs.set_item("project", project.clone().into_pyobject(py).unwrap())?;
            kwargs.set_item("domain", domain.clone().into_pyobject(py).unwrap())?;
            kwargs.set_item("run_name", run_name.clone().into_pyobject(py).unwrap())?;
            kwargs.set_item("name", action_name.clone().into_pyobject(py).unwrap())?;
            kwargs.set_item(
                "run_base_dir",
                run_base_dir.clone().into_pyobject(py).unwrap(),
            )?;
            // Set all the arguments from TaskArgs into the kwargs dictionary
            kwargs.set_item(
                "raw_data_path",
                args.raw_data_path.clone().into_pyobject(py).unwrap(),
            )?;
            kwargs.set_item(
                "output_path",
                args.outputs_path.clone().into_pyobject(py).unwrap(),
            )?;
            kwargs.set_item("version", args.version.clone().into_pyobject(py).unwrap())?;
            kwargs.set_item(
                "image_cache",
                args.image_cache.clone().into_pyobject(py).unwrap(),
            )?;
            kwargs.set_item(
                "checkpoint_path",
                args.checkpoint_path.clone().into_pyobject(py).unwrap(),
            )?;
            kwargs.set_item(
                "prev_checkpoint",
                args.prev_checkpoint.clone().into_pyobject(py).unwrap(),
            )?;
            kwargs.set_item("input_path", args.inputs.clone().into_pyobject(py).unwrap())?;
            let el = self.locals.event_loop(py);

            let coro = rusty.call_method("run_task", (), Some(&kwargs))?;
            let event_loop = self.locals.event_loop(py);
            let task = event_loop.call_method1("create_task", (coro,))?;
            let rust_fut = into_future_with_locals(&self.locals, task.clone())?;
            Ok::<_, PyErr>((task.unbind(), rust_fut))
        }).map_err(|e| TaskError::Python(e))?;

        let mut fut = Box::pin(fut);
        let result = tokio::select! {
            result = &mut fut => {
                debug!("Task future completed: {:?}", run_name);
                result.map_err(|e| TaskError::Python(e))
            }
            _ = token.cancelled() => {
                info!("Cancellation requested, canceling async Task for {:?}", run_name);
                cancel_python_task(&self.locals, &py_task).await?;

                let msg = format!("Task was aborted: run {:?}, action {:?}", run_name, action_name);
                Err(TaskError::Cancelled(msg))
            }
        };

        result
    }
}

async fn cancel_python_task(locals: &TaskLocals, py_task: &PyObject) -> Result<(), TaskError> {
    Python::with_gil(|py| -> PyResult<()> {
        let task = py_task.bind(py);
        let _ = task.call_method0("cancel")?;
        Ok(())
    }).map_err(|e| TaskError::Python(e))?;

    let wait_cancel_future = Python::with_gil(|py| {
        let task = py_task.bind(py);
        into_future_with_locals(locals, task.clone())
    }).map_err(|e| TaskError::Python(e))?;

    // todo rather than swallowing the wait_cancel_future error we should capture it
    if timeout(GRACE, wait_cancel_future).await.is_err() {
        // Grace period expired - this indicates a serious problem
        // Return a special error that should cause process termination
        warn!("Python task cancellation failed to finish, returning poison pill and terminating actor worker...");

        return Err(TaskError::CleanUpTimeoutPoisonPill(
            "Python task failed to clean up within grace period - process should terminate".to_string()
        ));
    }
    Ok(())
}

/// Creates a controller for task execution
#[instrument(fields(endpoint = ?endpoint, insecure))]
pub async fn create_controller(
    endpoint: Option<String>,
    insecure: bool,
    api_key: Option<String>,
) -> Result<Py<PyAny>, PyErr> {
    debug!("Creating controller");
    let fut = Python::with_gil(|py| {
        let rusty = py.import("flyte._internal.runtime.rusty")?;
        let kwargs = PyDict::new(py);
        kwargs.set_item("endpoint", endpoint.into_pyobject(py).unwrap())?;
        kwargs.set_item("insecure", insecure.into_pyobject(py).unwrap())?;
        kwargs.set_item("api_key", api_key.into_pyobject(py).unwrap())?;
        pyo3_async_runtimes::tokio::into_future(rusty.call_method(
            "create_controller",
            (),
            Some(&kwargs),
        )?)
    })?;

    let controller = fut.await?;
    info!("Controller created successfully");

    Ok(controller)
}
