use thiserror::Error;

use crate::{HelmCall as _, HelmCallImpl, SearchRequest, env::Env};

#[derive(Clone, Debug, Default)]
pub struct RepoSearch {
    pub versions: bool,
    pub regexp: String,
    pub devel: bool,
    pub version: String,
    pub terms: Vec<String>,
    pub env: Env,
}

impl From<RepoSearch> for SearchRequest {
    fn from(req: RepoSearch) -> Self {
        SearchRequest {
            versions: req.versions,
            regexp: req.regexp,
            devel: req.devel,
            version: req.version,
            terms: req.terms,
            env: req.env.into(),
        }
    }
}

#[derive(Error, Debug)]
pub enum RepoSearchError {
    #[error("repo search error: {err}")]
    RepoSearch {
        response: Option<String>,
        err: String,
    },
}

pub async fn repo_search(req: RepoSearch) -> Result<String, RepoSearchError> {
    let res = HelmCallImpl::repo_search(req.into()).await;
    if let Some(err) = res.0.err.first() {
        return Err(RepoSearchError::RepoSearch {
            response: match res.0.data.as_str() {
                "" => None,
                d => Some(d.to_string()),
            },
            err: err.clone(),
        });
    }

    Ok(res.0.data)
}
