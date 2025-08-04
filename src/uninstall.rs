use thiserror::Error;

use crate::{HelmCall as _, HelmCallImpl, UninstallRequest, env::Env};

#[derive(Clone, Debug)]
pub struct Uninstall {
    pub release_name: String,
    pub ns: String,
    pub disable_hooks: bool,
    pub dry_run: bool,
    pub ignore_not_found: bool,
    pub keep_history: bool,
    pub wait: bool,
    pub deletion_propagation: String,
    pub timeout: Vec<i64>,
    pub description: String,
    pub env: Env,
}

impl Default for Uninstall {
    fn default() -> Self {
        Uninstall {
            release_name: Default::default(),
            ns: Default::default(),
            disable_hooks: Default::default(),
            dry_run: Default::default(),
            ignore_not_found: Default::default(),
            keep_history: Default::default(),
            wait: Default::default(),
            deletion_propagation: Default::default(),
            timeout: vec![300],
            description: Default::default(),
            env: Default::default(),
        }
    }
}

impl From<Uninstall> for UninstallRequest {
    fn from(req: Uninstall) -> Self {
        UninstallRequest {
            release_name: req.release_name,
            ns: req.ns,
            disable_hooks: req.disable_hooks,
            dry_run: req.dry_run,
            ignore_not_found: req.ignore_not_found,
            keep_history: req.keep_history,
            wait: req.wait,
            deletion_propagation: req.deletion_propagation,
            timeout: req.timeout,
            description: req.description,
            env: req.env.into(),
        }
    }
}

#[derive(Error, Debug)]
pub enum UninstallError {
    #[error("uninstall error: {err}")]
    Uninstall {
        response: Option<String>,
        err: String,
    },
}

pub async fn uninstall(req: Uninstall) -> Result<String, UninstallError> {
    let res = HelmCallImpl::uninstall(req.into()).await;
    if let Some(err) = res.0.err.first() {
        return Err(UninstallError::Uninstall {
            response: match res.0.data.as_str() {
                "" => None,
                d => Some(d.to_string()),
            },
            err: err.clone(),
        });
    }

    Ok(res.0.data)
}
