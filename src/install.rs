use thiserror::Error;

use crate::{HelmCall as _, HelmCallImpl, InstallRequest, env::Env};

#[derive(Clone, Debug)]
pub struct Install {
    pub release_name: String,
    pub chart: String,
    pub version: String,
    pub ns: String,
    pub wait: bool,
    pub timeout: Vec<i64>,
    pub create_namespace: bool,
    pub values: Vec<u8>,
    pub dry_run: Option<String>,
    pub env: Env,
}

impl Default for Install {
    fn default() -> Self {
        Install {
            timeout: vec![300],
            release_name: Default::default(),
            chart: Default::default(),
            version: Default::default(),
            ns: Default::default(),
            wait: Default::default(),
            create_namespace: Default::default(),
            values: Default::default(),
            env: Default::default(),
            dry_run: Default::default(),
        }
    }
}

impl From<Install> for InstallRequest {
    fn from(req: Install) -> Self {
        InstallRequest {
            release_name: req.release_name,
            chart: req.chart,
            version: req.version,
            ns: req.ns,
            wait: req.wait,
            timeout: req.timeout,
            create_namespace: req.create_namespace,
            values: req.values,
            dry_run: req.dry_run.into_iter().collect(),
            env: req.env.into(),
        }
    }
}

#[derive(Error, Debug)]
pub enum InstallError {
    #[error("install error: {err}")]
    Install {
        response: Option<String>,
        err: String,
    },
}

pub async fn install(req: Install) -> Result<String, InstallError> {
    let res = HelmCallImpl::install(req.into()).await;
    if let Some(err) = res.0.err.first() {
        return Err(InstallError::Install {
            response: match res.0.data.as_str() {
                "" => None,
                d => Some(d.to_string()),
            },
            err: err.clone(),
        });
    }

    Ok(res.0.data)
}
