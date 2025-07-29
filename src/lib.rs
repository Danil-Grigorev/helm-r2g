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

#[derive(rust2go::R2G, Clone, Default)]
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
}
