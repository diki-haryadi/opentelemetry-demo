# Copyright The OpenTelemetry Authors
# SPDX-License-Identifier: Apache-2.0

variable "kubeconfig_path" {
  description = "Path to the kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "kube_context" {
  description = "Kubernetes context to use (leave empty for current context)"
  type        = string
  default     = ""
}

variable "namespace" {
  description = "Kubernetes namespace for OpenTelemetry Demo"
  type        = string
  default     = "otel-demo"
}

variable "release_name" {
  description = "Helm release name"
  type        = string
  default     = "opentelemetry-demo"
}

variable "chart_version" {
  description = "Version of the OpenTelemetry Demo Helm chart (leave empty for latest)"
  type        = string
  default     = ""
}

variable "service_type" {
  description = "Kubernetes service type for frontend proxy (ClusterIP, NodePort, LoadBalancer)"
  type        = string
  default     = "ClusterIP"
}

variable "image_pull_policy" {
  description = "Image pull policy for all containers"
  type        = string
  default     = "IfNotPresent"
}

variable "resource_limits_enabled" {
  description = "Enable resource limits for containers"
  type        = bool
  default     = false
}

variable "default_memory_limit" {
  description = "Default memory limit for containers"
  type        = string
  default     = "500Mi"
}

variable "default_cpu_limit" {
  description = "Default CPU limit for containers"
  type        = string
  default     = "500m"
}

variable "helm_values" {
  description = "Additional Helm values as YAML string"
  type        = string
  default     = ""
}

variable "custom_settings" {
  description = "Map of custom Helm chart settings"
  type        = map(string)
  default     = {}
}
