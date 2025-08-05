use helm_r2g::{
    Install, List, RegistryLogin, Uninstall, Upgrade, install, list, registry_login, uninstall,
    upgrade,
};
use serde_json::json;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let token = std::env::var("GITHUB_TOKEN").unwrap_or_default();
    let username = std::env::var("GITHUB_USER").unwrap_or_default();

    registry_login(RegistryLogin {
        hostname: "ghcr.io".to_string(),
        username: username.to_string(),
        password: token.to_string(),
        ..Default::default()
    })
    .await?;

    // Release name, chart and values
    let release_name = "helm-sdk-example";
    let chart_ref = "oci://ghcr.io/stefanprodan/charts/podinfo";
    let release_values = json!({
        "replicaCount": "2",
    });

    // Install the chart (from the pulled chart local archive)
    let req = Install {
        chart: chart_ref.to_string(),
        release_name: release_name.to_string(),
        wait: true,
        values: serde_json::to_vec(&release_values).unwrap(),
        ..Default::default()
    };

    // Run install
    install(req).await?;

    // List installed charts
    list(List {
        all_namespaces: true,
        ..Default::default()
    })
    .await?;

    // Upgrade to version 6.5.4, updating the replicaCount to three
    let release_values = json!({
        "replicaCount": "3",
    });

    let req = Upgrade {
        release_name: release_name.to_string(),
        chart: chart_ref.to_string(),
        version: "6.5.4".to_string(),
        wait: true,
        values: serde_json::to_vec(&release_values).unwrap(),
        ..Default::default()
    };

    // Run upgrade
    upgrade(req).await?;

    // List installed charts
    list(List {
        all_namespaces: true,
        ..Default::default()
    })
    .await?;

    // Uninstall the chart
    let req = Uninstall {
        release_name: release_name.to_string(),
        wait: true,
        ..Default::default()
    };

    // Run uninstall
    uninstall(req).await?;

    Ok(())
}
