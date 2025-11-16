# Copyright The OpenTelemetry Authors
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_version = ">= 1.0"

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

# Configure Kubernetes Provider
provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kube_context
}

# Configure Helm Provider
provider "helm" {
  kubernetes {
    config_path    = var.kubeconfig_path
    config_context = var.kube_context
  }
}

# Create namespace for OpenTelemetry Demo
resource "kubernetes_namespace" "otel_demo" {
  metadata {
    name = var.namespace
    labels = {
      name = var.namespace
      app  = "opentelemetry-demo"
    }
  }
}

# Deploy OpenTelemetry Demo using Helm
resource "helm_release" "opentelemetry_demo" {
  name       = var.release_name
  repository = "https://open-telemetry.github.io/opentelemetry-helm-charts"
  chart      = "opentelemetry-demo"
  version    = var.chart_version
  namespace  = kubernetes_namespace.otel_demo.metadata[0].name

  # Wait for all resources to be ready
  wait          = true
  wait_for_jobs = true
  timeout       = 600

  # Custom values can be set here
  values = var.helm_values != "" ? [var.helm_values] : []

  # Common configuration values
  set {
    name  = "default.enabled"
    value = "true"
  }

  set {
    name  = "default.image.pullPolicy"
    value = var.image_pull_policy
  }

  # Service type configuration
  set {
    name  = "frontendProxy.service.type"
    value = var.service_type
  }

  # Resource limits can be configured
  dynamic "set" {
    for_each = var.resource_limits_enabled ? [1] : []
    content {
      name  = "default.resources.limits.memory"
      value = var.default_memory_limit
    }
  }

  dynamic "set" {
    for_each = var.resource_limits_enabled ? [1] : []
    content {
      name  = "default.resources.limits.cpu"
      value = var.default_cpu_limit
    }
  }

  # Additional custom settings
  dynamic "set" {
    for_each = var.custom_settings
    content {
      name  = set.key
      value = set.value
    }
  }

  depends_on = [kubernetes_namespace.otel_demo]
}
