use thiserror::Error;

use crate::{HelmCall as _, HelmCallImpl, ListRequest, env::Env};

#[derive(Clone, Debug, Default)]
pub struct List {
    pub all: bool,
    pub all_namespaces: bool,
    pub sort: u64,
    pub by_date: bool,
    pub sort_reverse: bool,
    pub state_mask: u64,
    pub limit: i64,
    pub offset: i64,
    pub filter: String,
    pub no_headers: bool,
    pub time_format: String,
    pub uninstalled: bool,
    pub superseded: bool,
    pub uninstalling: bool,
    pub deployed: bool,
    pub failed: bool,
    pub pending: bool,
    pub selector: String,
    pub ns: String,
    pub env: Env,
}

impl From<List> for ListRequest {
    fn from(req: List) -> Self {
        ListRequest {
            all: req.all,
            all_namespaces: req.all_namespaces,
            sort: req.sort,
            by_date: req.by_date,
            sort_reverse: req.sort_reverse,
            state_mask: req.state_mask,
            limit: req.limit,
            offset: req.offset,
            filter: req.filter,
            no_headers: req.no_headers,
            time_format: req.time_format,
            uninstalled: req.uninstalled,
            superseded: req.superseded,
            uninstalling: req.uninstalling,
            deployed: req.deployed,
            failed: req.failed,
            pending: req.pending,
            selector: req.selector,
            ns: req.ns,
            env: req.env.into(),
        }
    }
}

#[derive(Error, Debug)]
pub enum ListError {
    #[error("list error: {err}")]
    List {
        response: Option<String>,
        err: String,
    },
}

pub async fn list(req: List) -> Result<String, ListError> {
    let res = HelmCallImpl::list(req.into()).await;
    if let Some(err) = res.0.err.first() {
        return Err(ListError::List {
            response: match res.0.data.as_str() {
                "" => None,
                d => Some(d.to_string()),
            },
            err: err.clone(),
        });
    }

    Ok(res.0.data)
}
