# Copyright The OpenTelemetry Authors
# SPDX-License-Identifier: Apache-2.0

output "namespace" {
  description = "The Kubernetes namespace where OpenTelemetry Demo is deployed"
  value       = kubernetes_namespace.otel_demo.metadata[0].name
}

output "release_name" {
  description = "The Helm release name"
  value       = helm_release.opentelemetry_demo.name
}

output "release_status" {
  description = "The status of the Helm release"
  value       = helm_release.opentelemetry_demo.status
}

output "chart_version" {
  description = "The version of the deployed Helm chart"
  value       = helm_release.opentelemetry_demo.version
}

output "access_instructions" {
  description = "Instructions to access the OpenTelemetry Demo"
  value       = <<-EOT
    OpenTelemetry Demo has been deployed successfully!

    To access the demo, you have several options depending on your service type:

    1. Port-forward (works for any service type):
       kubectl port-forward -n ${kubernetes_namespace.otel_demo.metadata[0].name} svc/${var.release_name}-frontendproxy 8080:8080

    Then access:
    - Demo UI: http://localhost:8080
    - Jaeger UI: http://localhost:8080/jaeger/ui
    - Grafana: http://localhost:8080/grafana/
    - Load Generator UI: http://localhost:8080/loadgen/
    - Feature Flags: http://localhost:8080/feature/

    2. Check all services:
       kubectl get svc -n ${kubernetes_namespace.otel_demo.metadata[0].name}

    3. Check all pods:
       kubectl get pods -n ${kubernetes_namespace.otel_demo.metadata[0].name}

    4. View logs for a specific service:
       kubectl logs -n ${kubernetes_namespace.otel_demo.metadata[0].name} -l app.kubernetes.io/component=<component-name>
  EOT
}
