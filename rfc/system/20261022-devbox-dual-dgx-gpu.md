# [RFC] Multi-node devbox with GPU/RDMA-capable AppEnvironments

**Authors:**

- @cosmicBboy

**Status:** Draft for comment

## 1 Executive Summary

Flyte 2's devbox is a **single-node** cluster (one privileged k3s container) and its
app-serving path **cannot hand an AppEnvironment a usable GPU**: the app controller drops
`Resources_GPU`, and pods that never request a GPU get no device nodes, no CUDA, and no
InfiniBand. This RFC proposes making devbox:

1. **span two physical DGX Spark (GB10) nodes** via a k3s agent joined over the existing
   RoCE fabric, and
2. give AppEnvironments **GPU + RDMA + shared-memory access on both nodes**, so a
   tensor-parallel (TP=2) DeepSeek-V4-Flash-0731 vLLM service can be hosted as a Flyte app
   instead of a systemd + docker-compose stack.

All three current blockers were confirmed live on hardware (evidence in §2). The change is a
small, well-scoped backend fix plus a devbox bring-up/ops layer; no wire-format changes.

## 2 Motivation

The runner of this RFC keeps a pair of NVIDIA DGX Sparks (`spark-a7e9` + `spark-17d7`),
connected by a 200 Gb RoCE fabric on `192.168.100.0/24`, serving
`deepseek-ai/DeepSeek-V4-Flash-0731` (155.4 GiB FP8; one 121 GiB GB10 cannot hold it) with
vLLM `--tensor-parallel-size 2 --nnodes 2`, DSpark speculative decoding, and NCCL all-reduce
over `rocep1s0f1` (GID 3). Today that service is supervised by a systemd user unit around a
docker-compose project. The motivation: run the same service as a Flyte 2 AppEnvironment on a
devbox that lives on the pair, replacing hand-rolled lifecycle glue with Flyte's app serving,
and make "one devbox, both Sparks" work for general multi-GPU app workloads.

### 2.1 Verified baseline (live test on the pair)

Flyte SDK 2.12.1, devbox `cr.flyte.org/flyteorg/flyte-devbox:gpu-latest` (arm64), run on
node 1 while the production service stayed healthy throughout:

| Observation | Result |
|---|---|
| Devbox on arm64/GB10 | boots; UI `:30080`; in-cluster registry `localhost:30000`; control plane ~1.9 GiB RAM |
| Cluster nodes | **1** (`16b6e02f1643`, k3s v1.34.6+k3s1); node 2 not schedulable |
| Node GPU allocatable | `nvidia.com/gpu: 1` (device-plugin DaemonSet healthy) |
| Node accelerator labels | none |
| AppEnvironment deploy | works (Knative service, scale-to-zero / min-scale) |
| Pod resources with `Resources(gpu="GB10:1")` | `{requests:{cpu:1,memory:1Gi}}` — **GPU dropped** |
| Pod GPU view | `/proc/driver/nvidia/gpus = 000f:01:00.0`, firmware 580.142, UUID = node-1 GB10 (`GPU-97e039b1-…`) — driver interface mounted, **no `/dev/nvidia*` nodes** |
| Pod `/dev` | no `/dev/nvidia*`, **no `/dev/infiniband`**, no `nvidia-smi`/CUDA |

A manual `kubectl label node … k8s.amazonaws.com/accelerator=GB10` satisfied node selection,
but the pod **still** received no `nvidia.com/gpu` — the defect is in the app backend
(§3.3), not the scheduler.

## 3 Proposed Implementation

### 3.1 Layer A — Multi-node devbox (both Sparks as k8s nodes)

The devbox is one `docker run --privileged … --gpus all <image> k3s server --cluster-init`
container (`flyte-sdk/src/flyte/cli/_devbox.py`, `docker/devbox-bundled/bin/k3d-entrypoint.sh`).
Node 2 joins as a **k3s agent** running the same `gpu-latest` image:

- `flyte-sdk/src/flyte/cli/_devbox.py`: add `--multi-node`. After the server is healthy,
  1. export the join token + API (`:6443`) **and the in-cluster registry (`:30000`)** on the
     fabric IPs so node 2 can reach them,
  2. `ssh 192.168.100.11 'docker run -d --privileged --gpus all … <image> k3s agent
     --server https://192.168.100.10:6443 --token <t>'`,
  3. ship the image to node 2 with the fabric-shipping pattern already used by the repo
     (`docker save | ssh -c aes128-gcm@openssh.com node2 docker load`), and
  4. merge the resulting nodes into the local kubeconfig.
- `docker/devbox-bundled/` `k3d-entrypoint.sh` already ends in `exec /bin/k3s "$@"`, so
  `k3s agent …` works from the same image **if** the agent also carries the nvidia containerd
  runtime config + NVIDIA Container Toolkit (it will: same `Dockerfile.gpu`).
- **Registry reachability (the detail that bites):** node 2's agent cannot see
  `localhost:30000` (that resolves to node 2's own docker). Publish the registry on
  `192.168.100.10:30000` and add it to the agent's containerd `insecure_registries`, or push
  app images to a registry both nodes can pull.
- Keep control-plane components on the server node (tolerations / `k3s.io` labels) so
  rustfs / postgres / Knative stay single-image; only workload nodes multiply.

### 3.2 Layer B — GPU devices that actually land in pods

Pods today see the driver interface but no `/dev/nvidia*` because **no pod ever requests a
GPU**: the CRI/kubelet then injects no devices, and the devbox's default `nvidia` containerd
runtime — per its own in-repo comment — is "a passthrough when no GPU is requested"
(`docker/devbox-bundled/containerd-config.toml.tmpl`). Plan:

- Land §3.3 first, then verify `/dev/nvidia*` appears when `nvidia.com/gpu` is requested.
- If injection still fails inside the nested (k3s-in-Docker) setup, switch to **CDI**:
  `nvidia-ctk cdi generate --output=/etc/cdi` in the devbox/agent container + containerd
  `enable_cdi = true`, triggered from the bootstrap.
- Confirm `docker/devbox-bundled/nvidia-device-plugin.yaml` runs on **both** nodes
  (DaemonSet) and that `nvcr.io/nvidia/k8s-device-plugin` handles GB10/arm64 (bump if not).

### 3.3 Layer C — AppEnvironment GPU/shared-memory plumbing (the backend fix)

The SDK already serializes `extended_resources` for apps
(`flyte-sdk/src/flyte/app/_runtime/app_serde.py:493`), including the accelerator and
shared-memory request; the backend ignores everything but Apple Silicon. Fix in
`flyte/app/internal/k8s/app_client.go`:

1. `protoResourceName`: map `Resources_GPU` → the configured accelerator resource
   (`nvidia.com/gpu` by default) instead of the current
   `default: return "", false` (which silently drops it).
2. `buildPodSpec` (`Spec_Container` case): call the task path's
   `flytek8s.SanitizeGPUResourceRequirements(&container.Resources, spec.GetExtendedResources().GetGpuAccelerator())`.
3. In `buildKService`/`buildPodSpec`: when a GPU accelerator is present, call
   `flytek8s.ApplyGPUNodeSelectors(&podSpec, accelerator)` (node selector
   `k8s.amazonaws.com/accelerator: GB10` + toleration — the normalized `GB10` label value
   comes out of the default `AcceleratorDevices` map since there is no built-in `GB10` entry;
   consider adding one).
4. Apply `ExtendedResources.SharedMemory` as a `/dev/shm` emptyDir volume (mirror the task
   path) so vLLM MPI/DSpark rendezvous buffers get real shared memory.
5. Auto-label nodes with the detected GPU
   (`kubectl label node … k8s.amazonaws.com/accelerator=GB10 nvidia.com/gpu=present`) in the
   devbox bring-up (bootstrap or CLI), so named-GPU selection matches without manual steps.

This single backend change unblocks GPU for every app type (FastAPI, vLLM, SGLang, …)

### 3.4 Layer D — RDMA, host networking, and memory for cross-node TP

- **InfiniBand:** no IB device plugin ships with Flyte. Minimal path: privileged
  pod-template apps mounting `hostPath: /dev/infiniband` (devbox containers are privileged, so
  the devices are available). Productized later: a small `flyte.dev/infiniband` DaemonSet
  device plugin so pods request it declaratively.
- **Host networking:** cross-node TP wants rank pods binding `192.168.100.10/.11` for the
  `MASTER_ADDR:25000` rendezvous. `PodTemplate` already supports `host_network=True`; the app
  build path must not strip it, and Knative hostNetwork implications need validating
  (kourier/queue-proxy). Fallback with zero platform change: ranks rendezvous over their pod
  IPs via fabric routes.
- **memlock / NCCL channels:** no k8s-level `--ulimit memlock=-1`. Document keeping the
  deployed NCCL env matrix (`NCCL_MAX_NCHANNELS=8` etc.) as pod env until the known node-2
  kernel/driver mismatch is reconciled — this is ops config, not code.

### 3.5 Layer E — Multi-host TP=2 serving pattern

`VLLMAppEnvironment` is one `vllm-fserve serve` process in one pod with
`--tensor-parallel-size N` meaning N GPUs **on one node**; it cannot span nodes. Two options:

- **Now:** deploy **two** AppEnvironments pinned by `kubernetes.io/hostname` node selector —
  `dspark-head` (rank 0) and `dspark-worker` (rank 1) — each `Scaling(replicas=(1,1))` so
  scale-to-zero cannot tear down one rank, using the GX10 image invoked directly
  (`--nnodes 2 --node-rank N --master-addr 192.168.100.10:25000` + NCCL env). Since the GX10
  image has no `fserve`, use a `PodTemplate` app (raw command, hostPath weights) rather than
  `VLLMAppEnvironment`.
- **Later:** a first-class multi-rank app (analogous to the MPI/Ray task plugins) where one
  AppEnvironment declares ranks and the controller materializes N matching revisions with
  `--node-rank`/`NNODES` plus a join service. New IDL + controller + SDK surface; optional.

### 3.6 Layer F — Shared weights on both nodes

Both ranks read the full checkpoint index. Two paths:

- **hostPath (no code):** mount each node's existing `~/.cache/huggingface` (node 2's copy is
  kept in sync today via rsync).
- **Streamed (productized):** `flyte.prefetch.hf_model(...)` → `model_path=RunOutput(...)`
  so each replica pulls weights from rustfs. Requires the `fserve` runtime in the image —
  pair with the "later" distributed app for a clean story.

### 3.7 Minimal-change path (works today, no backend GPU code)

A **PodTemplate AppEnvironment** bypasses §3.3 entirely because it supplies a raw k8s pod
spec: request `nvidia.com/gpu: 1` directly, mount `hostPath /dev/infiniband`, big
`/dev/shm`, `hostNetwork`, privileged, and run the GX10 image. Only §3.1 (multi-node devbox)
+ §3.2 (device injection on both nodes) are then required for a real cross-node GPU app;
§3.3 makes `Resources(gpu=…)` work the normal way afterward.

### 3.8 Change inventory

| Repo | Area | Change |
|---|---|---|
| flyte | `app/internal/k8s/app_client.go` | GPU → `nvidia.com/gpu`; `SanitizeGPUResourceRequirements`; `ApplyGPUNodeSelectors`; `/dev/shm` from `ExtendedResources.SharedMemory` |
| flyte | `docker/devbox-bundled/` | agent-capable entrypoint/config; per-node nvidia runtime + insecure registry; optional CDI bootstrap; IB hostPath docs or IB device-plugin manifest |
| flyte | bootstrap `cmd/bootstrap/main.go` | node auto-labeling + optional CDI generation |
| flyte-sdk | `flyte/cli/_devbox.py`, `_start.py` | `--multi-node`: token/port/registry wiring, ssh `k3s agent` bring-up, kubeconfig merge, node labeling |
| flyte-sdk | `plugins/vllm` | (optional) GX10-compatible image + multi-host `VLLMAppEnvironment` |
| flyte | (future) app IDL + controller | first-class multi-rank app (§3.5 "later") |

## 4 Metrics & Dashboards

- Devbox cluster node count (`kubectl get nodes`) — must be 2 on the pair.
- App pod GPU outcome: `nvidia.com/gpu` requested, `/dev/nvidia0` present, per-node GPU UUID
  distinct (`/proc/driver/nvidia/gpus`).
- Capability gate: in-pod `ib_write_bw` / 2-rank NCCL all-reduce succeeds over RoCE
  (RDMA, not TCP).
- Serving gate: TP=2 smoke model moves `free -g` on **both** nodes together (existing
  TP-spanning signature).

## 5 Drawbacks

- Complexity: a multi-node devbox is more moving parts than today's single container; the
  current systemd + docker-compose service is simpler and already proven (~75 tok/s).
- Knative serving was not designed for hostNetwork/privileged long-running ranks; some of
  §3.4 may require accommodation or a non-Knative app path.
- The devbox remains a local cluster; this RFC does not make it a production control plane.

## 6 Alternatives

1. **Do nothing, keep systemd + docker-compose.** Lowest complexity; zero Flyte value gained
   for this workload; no access to Flyte job/batch orchestration next to the model.
2. **Full multi-node Flyte (not devbox)** using the `flyte-binary` Helm chart with node 2 as a
   k3s agent. Same layers A–F apply; heavier than devbox, needed only if production serving is
   the goal.
3. **Plain tasks instead of apps** for the serving pods. Tasks already get GPU plumbing; but
   tasks are batch-oriented (ephemeral, no durable HTTP endpoint), so not suited to an
   always-on model endpoint.
4. **Leverage an existing multi-node distributed framework** (MPIOperator/Ray job already in
   Flyte for tasks) as the serving substrate for a "later" multi-rank app — the natural
   long-term home but more work than this RFC's immediate fix.

## 7 Potential Impact and Dependencies

- **Scope of impact:** the devbox image/build (new agent mode), the app controller resource
  handling (small, additive), and devbox node labeling. No wire-format, no task-path changes.
- **Security:** privileged/hostNetwork pods are powerful; the IB/privileged escape hatch
  must be opt-in per app (PodTemplate/`allow_fuse(privileged=True)` style), gated by cluster
  policy rather than defaulted.
- **Reproducibility:** pin the agent image to the same digest as the server to avoid the
  already-documented node-2 driver/kernel skew affecting cross-node NCCL.

## 8 Unresolved questions

- Does device injection work end-to-end once a pod *actually requests* `nvidia.com/gpu` in
  nested k3s, or is a CDI fallback required? (First experiment to run — see §3.2.)
- Which Knative knobs (if any) are needed for hostNetwork + privileged serving pods, and are
  they safe to default off?
- Should `GB10` be added to the built-in `AcceleratorDevices` mapping
  (`flyteplugins/…/flytek8s/config/config.go`) so labeled values match conventions used by the
  broader platform?
- Is a first-class multi-rank app (§3.5 "later") in-scope for the next milestone, or is the
  two-pinned-Apps pattern sufficient?

## 9 Conclusion

The motivating dual-DGX workload cannot run on a Flyte devbox today for exactly three,
already-isolated reasons: devbox is single-node, the app path drops GPU requests, and pods
have no RDMA. Each has a small, well-scoped fix; none requires a wire-format change. Landing
Layers A–C opens the devbox to both Sparks' GPUs (and, once verified, makes the Layer C
single-file backend fix the linchpin); Layers D–F turn that plumbing into a working TP=2
serving app. The minimal-change path lets early adopters validate the full end-to-end story on
hostPath + pod-template apps before the platform plumbing lands.

## 10 Prior Art (repo-internal references)

- Devbox image & runtime: `docker/devbox-bundled/` (`Dockerfile`, `Dockerfile.gpu`,
  `containerd-config.toml.tmpl`, `nvidia-device-plugin.yaml`, `bin/k3d-entrypoint*.sh`,
  `bootstrap/cmd/bootstrap/main.go`).
- App controller: `app/internal/k8s/app_client.go` (`buildPodSpec`, `protoResourceName`,
  `buildKService`).
- Task-path GPU helpers to reuse: `flyteplugins/go/tasks/pluginmachinery/flytek8s/pod_helper.go`
  (`SanitizeGPUResourceRequirements`, `ApplyGPUNodeSelectors`), config defaults +
  `AcceleratorDevices` in `…/flytek8s/config/config.go`.
- SDK serialization (already correct): `flyte-sdk/src/flyte/app/_runtime/app_serde.py`;
  `flyte-sdk/src/flyte/_pod.py` (`PodTemplate`), `plugins/vllm/…/vllm/_app_environment.py`.
