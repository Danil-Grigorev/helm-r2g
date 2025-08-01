use helm_r2g::{
    HelmCall as _, HelmCallImpl, InstallRequest, ListRequest,
    UninstallRequest, UpgradeRequest, LoginRequest,
};
use serde_json::json;

#[tokio::main]
async fn main() {
    let token = std::env::var("GITHUB_TOKEN").unwrap_or_default();
    let username = std::env::var("GITHUB_USER").unwrap_or_default();

    let login_req = HelmCallImpl::registry_login(LoginRequest {
        hostname: "ghcr.io".to_string(),
        username: username.to_string(),
        password: token.to_string(),
        ..Default::default()
    })
    .await;
    if !login_req.0.err.is_empty() {
        println!("failed to run registry login: {:?}", login_req.0.err);
    }

    // Release name, chart and values
    let release_name = "helm-sdk-example";
    let chart_ref = "oci://ghcr.io/stefanprodan/charts/podinfo";
    let release_values = json!({
        "replicaCount": "2",
    });

    // Install the chart (from the pulled chart local archive)
    let req = InstallRequest {
        chart: chart_ref.to_string(),
        release_name: release_name.to_string(),
        wait: true,
        values: serde_json::to_vec(&release_values).unwrap(),
        ..Default::default()
    };

    // Run install
    let res = HelmCallImpl::install(req).await;
    if !res.0.err.is_empty() {
        println!("failed to run install: {:?}", res.0.err);
        std::process::exit(1);
    }

    // List installed charts
    let res = HelmCallImpl::list(ListRequest {
        all_namespaces: true,
        ..Default::default()
    })
    .await;
    if !res.0.err.is_empty() {
        println!("failed to run list: {:?}", res.0.err);
        std::process::exit(1);
    }

    // Upgrade to version 6.5.4, updating the replicaCount to three
    let release_values = json!({
        "replicaCount": "3",
    });

    let req = UpgradeRequest {
        release_name: release_name.to_string(),
        chart: chart_ref.to_string(),
        version: "6.5.4".to_string(),
        wait: true,
        values: serde_json::to_vec(&release_values).unwrap(),
        ..Default::default()
    };

    // Run upgrade
    let res = HelmCallImpl::upgrade(req).await;
    if !res.0.err.is_empty() {
        println!("failed to run upgrade: {:?}", res.0.err);
        std::process::exit(1);
    }

    // List installed charts
    let res = HelmCallImpl::list(ListRequest {
        all_namespaces: true,
        ..Default::default()
    })
    .await;
    if !res.0.err.is_empty() {
        println!("failed to run list: {:?}", res.0.err);
        std::process::exit(1);
    }

    // Uninstall the chart
    let req = UninstallRequest {
        release_name: release_name.to_string(),
        wait: true,
        ..Default::default()
    };

    // Run uninstall
    let res = HelmCallImpl::uninstall(req).await;
    if !res.0.err.is_empty() {
        println!("failed to run uninstall: {:?}", res.0.err);
        std::process::exit(1);
    }
}
