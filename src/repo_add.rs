use thiserror::Error;

use crate::{AddRequest, HelmCall as _, HelmCallImpl, env::Env};

#[derive(Clone, Debug, Default)]
pub struct RepoAdd {
    pub name: String,
    pub url: String,
    pub username: String,
    pub password: String,
    pub password_from_stdin_opt: bool,
    pub pass_credentials_all: bool,
    pub force_update: bool,
    pub allow_deprecated_repos: bool,
    pub cert_file: String,
    pub key_file: String,
    pub ca_file: String,
    pub insecure_skip_tls_sverify: bool,
    pub env: Env,
}

impl From<RepoAdd> for AddRequest {
    fn from(req: RepoAdd) -> Self {
        AddRequest {
            name: req.name,
            url: req.url,
            username: req.username,
            password: req.password,
            password_from_stdin_opt: req.password_from_stdin_opt,
            pass_credentials_all: req.pass_credentials_all,
            force_update: req.force_update,
            allow_deprecated_repos: req.allow_deprecated_repos,
            cert_file: req.cert_file,
            key_file: req.key_file,
            ca_file: req.ca_file,
            insecure_skip_tls_sverify: req.insecure_skip_tls_sverify,
            env: req.env.into(),
        }
    }
}

#[derive(Error, Debug)]
pub enum RepoAddError {
    #[error("repo add error: {err}")]
    RepoAdd {
        response: Option<String>,
        err: String,
    },
}

pub async fn repo_add(req: RepoAdd) -> Result<(), RepoAddError> {
    let res = HelmCallImpl::repo_add(req.into()).await;
    if let Some(err) = res.0.err.first() {
        return Err(RepoAddError::RepoAdd {
            response: None, // AddResponse does not have a data field
            err: err.clone(),
        });
    }

    Ok(())
}
