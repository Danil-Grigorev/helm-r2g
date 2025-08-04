use crate::HelmEnv;

#[derive(Clone, Default, Debug)]
pub struct Env {
    // KubeConfig is the path to the kubeconfig file
    pub kube_config: Option<String>,
    // KubeContext is the name of the kubeconfig context.
    pub kube_context: Option<String>,
    // Bearer KubeToken used for authentication
    pub kube_token: Option<String>,
    // Custom certificate authority file.
    pub kube_ca_file: Option<String>,
    // KubeInsecureSkipTLSVerify indicates if server's certificate will not be checked for validity.
    // This makes the HTTPS connections insecure
    pub kube_insecure_skip_tls_verify: bool,
}

impl From<Env> for HelmEnv {
    fn from(value: Env) -> Self {
        HelmEnv {
            kube_config: value.kube_config.into_iter().collect(),
            kube_context: value.kube_context.into_iter().collect(),
            kube_token: value.kube_token.into_iter().collect(),
            kube_ca_file: value.kube_ca_file.into_iter().collect(),
            kube_insecure_skip_tls_verify: value.kube_insecure_skip_tls_verify,
        }
    }
}
