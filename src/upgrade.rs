use thiserror::Error;

use crate::{HelmCall as _, HelmCallImpl, UpgradeRequest, env::Env};

#[derive(Clone, Debug)]
pub struct Upgrade {
    pub release_name: String,
    pub chart: String,
    pub version: String,
    pub ns: String,
    pub wait: bool,
    pub timeout: Vec<i64>,
    pub dry_run: Option<String>,
    pub reuse_values: bool,
    pub reset_values: bool,
    pub values: Vec<u8>,
    pub env: Env,
}

impl Default for Upgrade {
    fn default() -> Self {
        Upgrade {
            timeout: vec![300],
            release_name: Default::default(),
            chart: Default::default(),
            version: Default::default(),
            ns: Default::default(),
            wait: Default::default(),
            dry_run: Default::default(),
            reuse_values: Default::default(),
            reset_values: Default::default(),
            values: Default::default(),
            env: Default::default(),
        }
    }
}

impl From<Upgrade> for UpgradeRequest {
    fn from(req: Upgrade) -> Self {
        UpgradeRequest {
            release_name: req.release_name,
            chart: req.chart,
            version: req.version,
            ns: req.ns,
            wait: req.wait,
            timeout: req.timeout,
            dry_run: req.dry_run.into_iter().collect(),
            reuse_values: req.reuse_values,
            reset_values: req.reset_values,
            values: req.values,
            env: req.env.into(),
        }
    }
}

#[derive(Error, Debug)]
pub enum UpgradeError {
    #[error("upgrade error: {err}")]
    Upgrade {
        response: Option<String>,
        err: String,
    },
}

pub async fn upgrade(req: Upgrade) -> Result<String, UpgradeError> {
    let res = HelmCallImpl::upgrade(req.into()).await;
    if let Some(err) = res.0.err.first() {
        return Err(UpgradeError::Upgrade {
            response: match res.0.data.as_str() {
                "" => None,
                d => Some(d.to_string()),
            },
            err: err.clone(),
        });
    }

    Ok(res.0.data)
}
