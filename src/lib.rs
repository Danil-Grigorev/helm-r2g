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

#[derive(rust2go::R2G, Clone)]
pub struct HelmChartInstallRequest {
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

impl Default for HelmChartInstallRequest {
    fn default() -> Self {
        HelmChartInstallRequest {
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

#[derive(rust2go::R2G, Clone, Default)]
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

#[derive(rust2go::R2G, Clone)]
pub struct HelmChartInstallResponse {
    pub err: Vec<String>,
    pub data: String,
}

#[derive(rust2go::R2G, Clone, Default)]
pub struct HelmChartListRequest {
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

#[derive(rust2go::R2G, Clone)]
pub struct HelmChartListResponse {
    pub err: Vec<String>,
    pub data: String,
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
    async fn install(req: HelmChartInstallRequest) -> HelmChartInstallResponse;
    #[drop_safe_ret]
    async fn list(req: HelmChartListRequest) -> HelmChartListResponse;
}
