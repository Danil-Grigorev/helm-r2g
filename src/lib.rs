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

#[derive(rust2go::R2G, Clone, Debug)]
pub struct InstallRequest {
    pub release_name: String,
    pub chart: String,
    pub version: String,
    pub ns: String,
    pub wait: bool,
    pub timeout: Vec<i64>,
    pub create_namespace: bool,
    pub values: Vec<u8>,
    pub env: HelmEnv,
    pub dry_run: Vec<String>,
}

impl Default for InstallRequest {
    fn default() -> Self {
        InstallRequest {
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

#[derive(rust2go::R2G, Clone, Debug)]
pub struct UpgradeRequest {
    pub release_name: String,
    pub chart: String,
    pub version: String,
    pub ns: String,
    pub wait: bool,
    pub timeout: Vec<i64>,
    pub values: Vec<u8>,
    pub env: HelmEnv,
    pub reset_values: bool,
    pub reuse_values: bool,
    pub dry_run: Vec<String>,
}

impl Default for UpgradeRequest {
    fn default() -> Self {
        UpgradeRequest {
            timeout: vec![300],
            release_name: Default::default(),
            chart: Default::default(),
            version: Default::default(),
            ns: Default::default(),
            wait: Default::default(),
            values: Default::default(),
            env: Default::default(),
            dry_run: Default::default(),
            reset_values: Default::default(),
            reuse_values: Default::default(),
        }
    }
}

#[derive(rust2go::R2G, Clone, Default, Debug)]
pub struct HelmEnv {
    // KubeConfig is the path to the kubeconfig file
    pub kube_config: Vec<String>,
    // KubeContext is the name of the kubeconfig context.
    pub kube_context: Vec<String>,
    // Bearer KubeToken used for authentication
    pub kube_token: Vec<String>,
    // Custom certificate authority file.
    pub kube_ca_file: Vec<String>,
    // KubeInsecureSkipTLSVerify indicates if server's certificate will not be checked for validity.
    // This makes the HTTPS connections insecure
    pub kube_insecure_skip_tls_verify: bool,
}

#[derive(rust2go::R2G, Clone, Debug)]
pub struct InstallResponse {
    pub err: Vec<String>,
    pub data: String,
}

#[derive(rust2go::R2G, Clone, Debug)]
pub struct UpgradeResponse {
    pub err: Vec<String>,
    pub data: String,
}

#[derive(rust2go::R2G, Clone, Default, Debug)]
pub struct ListRequest {
    pub ns: String,
    pub env: HelmEnv,
    // All ignores the limit/offset
    pub all: bool,
    // AllNamespaces searches across namespaces
    pub all_namespaces: bool,
    // Sort indicates the sort to use
    //
    // see pkg/releaseutil for several useful sorters
    pub sort: u64,
    // Overrides the default lexicographic sorting
    pub by_date: bool,
    pub sort_reverse: bool,
    // StateMask accepts a bitmask of states for items to show.
    // The default is ListDeployed
    pub state_mask: u64,
    // Limit is the number of items to return per Run()
    pub limit: i64,
    // Offset is the starting index for the Run() call
    pub offset: i64,
    // Filter is a filter that is applied to the results
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
}

#[derive(rust2go::R2G, Clone, Debug)]
pub struct ListResponse {
    pub err: Vec<String>,
    pub data: String,
}

#[derive(rust2go::R2G, Clone, Default, Debug)]
pub struct SearchRequest {
    pub versions: bool,
    pub regexp: String,
    pub devel: bool,
    pub version: String,
    pub terms: Vec<String>,
    pub env: HelmEnv,
}

#[derive(rust2go::R2G, Clone, Debug)]
pub struct SearchResponse {
    pub err: Vec<String>,
    pub data: String,
}

#[derive(rust2go::R2G, Clone, Default, Debug)]
pub struct AddRequest {
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

    pub env: HelmEnv,
}

#[derive(rust2go::R2G, Clone, Debug)]
pub struct AddResponse {
    pub err: Vec<String>,
}

#[derive(rust2go::R2G, Clone, Debug)]
pub struct UninstallRequest {
    pub ns: String,
    pub release_name: String,
    pub disable_hooks: bool,
    pub dry_run: bool,
    pub ignore_not_found: bool,
    pub keep_history: bool,
    pub wait: bool,
    pub deletion_propagation: String,
    pub timeout: Vec<i64>,
    pub description: String,

    pub env: HelmEnv,
}

impl Default for UninstallRequest {
    fn default() -> Self {
        UninstallRequest {
            timeout: vec![300],
            ns: Default::default(),
            release_name: Default::default(),
            disable_hooks: Default::default(),
            dry_run: Default::default(),
            ignore_not_found: Default::default(),
            keep_history: Default::default(),
            wait: Default::default(),
            deletion_propagation: Default::default(),
            description: Default::default(),
            env: Default::default(),
        }
    }
}

#[derive(rust2go::R2G, Clone, Debug)]
pub struct UninstallResponse {
    pub err: Vec<String>,
    pub data: String,
}

#[derive(rust2go::R2G, Clone, Debug, Default)]
pub struct LoginRequest {
    pub hostname: String,
    pub username: String,
    pub password: String,
    pub cert_file: String,
    pub key_file: String,
    pub ca_file: String,
    pub insecure: bool,
    pub plain_http: bool,

    pub env: HelmEnv,
}

#[derive(rust2go::R2G, Clone, Debug)]
pub struct LoginResponse {
    pub err: Vec<String>,
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
pub trait HelmCall {
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
