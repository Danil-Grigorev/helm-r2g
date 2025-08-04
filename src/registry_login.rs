use thiserror::Error;

use crate::{HelmCall as _, HelmCallImpl, LoginRequest, env::Env};

#[derive(Clone, Debug, Default)]
pub struct RegistryLogin {
    pub hostname: String,
    pub username: String,
    pub password: String,
    pub cert_file: String,
    pub key_file: String,
    pub ca_file: String,
    pub insecure: bool,
    pub plain_http: bool,
    pub env: Env,
}

impl From<RegistryLogin> for LoginRequest {
    fn from(req: RegistryLogin) -> Self {
        LoginRequest {
            hostname: req.hostname,
            username: req.username,
            password: req.password,
            cert_file: req.cert_file,
            key_file: req.key_file,
            ca_file: req.ca_file,
            insecure: req.insecure,
            plain_http: req.plain_http,
            env: req.env.into(),
        }
    }
}

#[derive(Error, Debug)]
pub enum RegistryLoginError {
    #[error("registry login error: {err}")]
    RegistryLogin {
        err: String,
    },
}

pub async fn registry_login(req: RegistryLogin) -> Result<(), RegistryLoginError> {
    let res = HelmCallImpl::registry_login(req.into()).await;
    if let Some(err) = res.0.err.first() {
        return Err(RegistryLoginError::RegistryLogin {
            err: err.clone(),
        });
    }

    Ok(())
}
