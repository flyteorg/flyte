#!/bin/sh

set -o errexit
set -o nounset

FLYTE_SANDBOX_GPU="${FLYTE_SANDBOX_GPU:-false}"

if [ "$FLYTE_SANDBOX_GPU" != "true" ]; then
  echo "[$(date -Iseconds)] [GPU] GPU support not enabled (FLYTE_SANDBOX_GPU=$FLYTE_SANDBOX_GPU)"
  exit 0
fi

echo "[$(date -Iseconds)] [GPU] Configuring NVIDIA GPU support..."

# Configure K3s containerd to use the NVIDIA container runtime.
# K3s picks up containerd config templates from this path.
mkdir -p /var/lib/rancher/k3s/agent/etc/containerd
cat > /var/lib/rancher/k3s/agent/etc/containerd/config.toml.tmpl <<'EOF'
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes."nvidia"]
  privileged_without_host_devices = false
  runtime_engine = ""
  runtime_root = ""
  runtime_type = "io.containerd.runc.v2"

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes."nvidia".options]
  BinaryName = "/usr/bin/nvidia-container-runtime"

[plugins."io.containerd.grpc.v1.cri".containerd]
  default_runtime_name = "nvidia"
EOF

# Deploy the NVIDIA device plugin as a K3s auto-deploy manifest.
cat > /var/lib/rancher/k3s/server/manifests/nvidia-device-plugin.yaml <<'EOF'
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: nvidia-device-plugin-daemonset
  namespace: kube-system
  labels:
    app.kubernetes.io/name: nvidia-device-plugin
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: nvidia-device-plugin
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        app.kubernetes.io/name: nvidia-device-plugin
    spec:
      tolerations:
        - key: nvidia.com/gpu
          operator: Exists
          effect: NoSchedule
      priorityClassName: system-node-critical
      containers:
        - name: nvidia-device-plugin
          image: nvcr.io/nvidia/k8s-device-plugin:v0.17.0
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
          volumeMounts:
            - name: device-plugin
              mountPath: /var/lib/kubelet/device-plugins
      volumes:
        - name: device-plugin
          hostPath:
            path: /var/lib/kubelet/device-plugins
EOF

echo "[$(date -Iseconds)] [GPU] NVIDIA GPU support configured"
