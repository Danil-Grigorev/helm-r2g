// Include the binding file. There are 2 ways to include it:
// 1. Use rust2go's macro:
// ```rust
// pub mod binding {
//     rust2go::r2g_include_binding!();
// }
// ```
// 2. Include it manually:
// ```rust
// pub mod binding {
//     include!(concat!(env!("OUT_DIR"), "/_go_bindings.rs"));
// }
// ```
// If you want to use your own binding file name, use:
// `rust2go::r2g_include_binding!("_go_bindings.rs");`
pub mod binding {
    #![allow(warnings)]
    rust2go::r2g_include_binding!();
}

pub mod env;
pub mod install;
pub mod list;
pub mod registry_login;
pub mod repo_add;
pub mod repo_search;
pub mod uninstall;
pub mod upgrade;

pub use env::Env;
pub use install::{Install, InstallError, install};
pub use list::{List, ListError, list};
pub use registry_login::{RegistryLogin, RegistryLoginError, registry_login};
pub use repo_add::{RepoAdd, RepoAddError, repo_add};
pub use repo_search::{RepoSearch, RepoSearchError, repo_search};
pub use uninstall::{Uninstall, UninstallError, uninstall};
pub use upgrade::{Upgrade, UpgradeError, upgrade};

#[derive(rust2go::R2G)]
struct InstallRequest {
    release_name: String,
    chart: String,
    version: String,
    ns: String,
    wait: bool,
    timeout: Vec<i64>,
    create_namespace: bool,
    values: Vec<u8>,
    env: HelmEnv,
    dry_run: Vec<String>,
}

#[derive(rust2go::R2G)]
struct UpgradeRequest {
    release_name: String,
    chart: String,
    version: String,
    ns: String,
    wait: bool,
    timeout: Vec<i64>,
    values: Vec<u8>,
    env: HelmEnv,
    reset_values: bool,
    reuse_values: bool,
    dry_run: Vec<String>,
}

#[derive(rust2go::R2G)]
struct HelmEnv {
    // KubeConfig is the path to the kubeconfig file
    kube_config: Vec<String>,
    // KubeContext is the name of the kubeconfig context.
    kube_context: Vec<String>,
    // Bearer KubeToken used for authentication
    kube_token: Vec<String>,
    // Custom certificate authority file.
    kube_ca_file: Vec<String>,
    // KubeInsecureSkipTLSVerify indicates if server's certificate will not be checked for validity.
    // This makes the HTTPS connections insecure
    kube_insecure_skip_tls_verify: bool,
}

#[derive(rust2go::R2G)]
struct InstallResponse {
    err: Vec<String>,
    data: String,
}

#[derive(rust2go::R2G)]
struct UpgradeResponse {
    err: Vec<String>,
    data: String,
}

#[derive(rust2go::R2G)]
struct ListRequest {
    ns: String,
    env: HelmEnv,
    // All ignores the limit/offset
    all: bool,
    // AllNamespaces searches across namespaces
    all_namespaces: bool,
    // Sort indicates the sort to use
    //
    // see pkg/releaseutil for several useful sorters
    sort: u64,
    // Overrides the default lexicographic sorting
    by_date: bool,
    sort_reverse: bool,
    // StateMask accepts a bitmask of states for items to show.
    // The default is ListDeployed
    state_mask: u64,
    // Limit is the number of items to return per Run()
    limit: i64,
    // Offset is the starting index for the Run() call
    offset: i64,
    // Filter is a filter that is applied to the results
    filter: String,
    no_headers: bool,
    time_format: String,
    uninstalled: bool,
    superseded: bool,
    uninstalling: bool,
    deployed: bool,
    failed: bool,
    pending: bool,
    selector: String,
}

#[derive(rust2go::R2G)]
struct ListResponse {
    err: Vec<String>,
    data: String,
}

#[derive(rust2go::R2G)]
struct SearchRequest {
    versions: bool,
    regexp: String,
    devel: bool,
    version: String,
    terms: Vec<String>,
    env: HelmEnv,
}

#[derive(rust2go::R2G)]
struct SearchResponse {
    err: Vec<String>,
    data: String,
}

#[derive(rust2go::R2G)]
struct AddRequest {
    name: String,
    url: String,
    username: String,
    password: String,
    password_from_stdin_opt: bool,
    pass_credentials_all: bool,
    force_update: bool,
    allow_deprecated_repos: bool,
    cert_file: String,
    key_file: String,
    ca_file: String,
    insecure_skip_tls_sverify: bool,

    env: HelmEnv,
}

#[derive(rust2go::R2G)]
struct AddResponse {
    err: Vec<String>,
}

#[derive(rust2go::R2G)]
struct UninstallRequest {
    ns: String,
    release_name: String,
    disable_hooks: bool,
    dry_run: bool,
    ignore_not_found: bool,
    keep_history: bool,
    wait: bool,
    deletion_propagation: String,
    timeout: Vec<i64>,
    description: String,

    env: HelmEnv,
}

#[derive(rust2go::R2G)]
struct UninstallResponse {
    err: Vec<String>,
    data: String,
}

#[derive(rust2go::R2G)]
struct LoginRequest {
    hostname: String,
    username: String,
    password: String,
    cert_file: String,
    key_file: String,
    ca_file: String,
    insecure: bool,
    plain_http: bool,

    env: HelmEnv,
}

#[derive(rust2go::R2G)]
struct LoginResponse {
    err: Vec<String>,
}

// Define the call trait.
// It can be defined in 2 styles: sync and async.
// If the golang side is purely calculation logic, and not very heavy, use sync can be more efficient.
// Otherwise, use async style:
// Both `async fn` and `impl Future` styles are supported.
//
// If you want to use your own binding mod name, use:
// `#[rust2go::r2g(binding)]`
#[rust2go::r2g]
trait HelmCall {
    #[drop_safe_ret]
    async fn install(req: InstallRequest) -> InstallResponse;
    #[drop_safe_ret]
    async fn upgrade(req: UpgradeRequest) -> UpgradeResponse;
    #[drop_safe_ret]
    async fn uninstall(req: UninstallRequest) -> UninstallResponse;
    #[drop_safe_ret]
    async fn list(req: ListRequest) -> ListResponse;
    #[drop_safe_ret]
    async fn repo_add(req: AddRequest) -> AddResponse;
    #[drop_safe_ret]
    async fn repo_search(req: SearchRequest) -> SearchResponse;
    #[drop_safe_ret]
    async fn registry_login(req: LoginRequest) -> LoginResponse;
}
