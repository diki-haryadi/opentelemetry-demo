terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
  }
}

provider "kubernetes" {
  config_path = "~/.kube/config"
}

# Namespace
resource "kubernetes_namespace" "rctiplus" {
  metadata {
    name = var.namespace
    labels = {
      name = var.namespace
    }
  }
}

# ConfigMaps
resource "kubernetes_config_map" "tempo_config" {
  metadata {
    name      = "tempo-config"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  data = {
    "tempo.yaml" = file("${path.module}/../infrastructure/tempo/tempo.yaml")
  }
}

resource "kubernetes_config_map" "alloy_config" {
  metadata {
    name      = "alloy-config"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  data = {
    "config.alloy" = file("${path.module}/../infrastructure/alloy/config.alloy")
  }
}

resource "kubernetes_config_map" "prometheus_config" {
  metadata {
    name      = "prometheus-config"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  data = {
    "prometheus.yml" = file("${path.module}/../infrastructure/mimir/prometheus.yml")
  }
}

resource "kubernetes_config_map" "mimir_config" {
  metadata {
    name      = "mimir-config"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  data = {
    "mimir.yaml" = file("${path.module}/../infrastructure/mimir/mimir.yaml")
  }
}

resource "kubernetes_config_map" "loki_config" {
  metadata {
    name      = "loki-config"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  data = {
    "loki.yaml" = file("${path.module}/../infrastructure/loki/loki.yaml")
  }
}

resource "kubernetes_config_map" "grafana_datasources" {
  metadata {
    name      = "grafana-datasources"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  data = {
    "datasources.yml" = file("${path.module}/../infrastructure/grafana/datasources.yml")
  }
}

resource "kubernetes_config_map" "grafana_dashboards" {
  metadata {
    name      = "grafana-dashboards"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  data = {
    "dashboards.yml" = file("${path.module}/../infrastructure/grafana/dashboards.yml")
  }
}

resource "kubernetes_config_map" "grafana_dashboard_files" {
  metadata {
    name      = "grafana-dashboard-files"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  data = {
    "service-tree.json"    = file("${path.module}/../infrastructure/grafana/dashboards/service-tree.json")
    "infrastructure.json"  = file("${path.module}/../infrastructure/grafana/dashboards/infrastructure.json")
    "red-metrics.json"     = file("${path.module}/../infrastructure/grafana/dashboards/red-metrics.json")
  }
}

# Persistent Volume Claims
resource "kubernetes_persistent_volume_claim" "tempo_data" {
  metadata {
    name      = "tempo-data"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "10Gi"
      }
    }
  }
}

resource "kubernetes_persistent_volume_claim" "prometheus_data" {
  metadata {
    name      = "prometheus-data"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "10Gi"
      }
    }
  }
}

resource "kubernetes_persistent_volume_claim" "mimir_data" {
  metadata {
    name      = "mimir-data"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "10Gi"
      }
    }
  }
}

resource "kubernetes_persistent_volume_claim" "loki_data" {
  metadata {
    name      = "loki-data"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "10Gi"
      }
    }
  }
}

resource "kubernetes_persistent_volume_claim" "grafana_data" {
  metadata {
    name      = "grafana-data"
    namespace = kubernetes_namespace.rctiplus.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "5Gi"
      }
    }
  }
}

# Deployments and Services
# Note: For actual deployment, you can use kubectl apply with the YAML files in kubernetes/ directory
# or use terraform kubernetes_manifest resources. The YAML files use Helm templates,
# so they need to be processed with Helm or templated first.
