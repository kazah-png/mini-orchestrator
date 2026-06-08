<div align="center">

# Mini Orchestrator — Container Orchestrator from Scratch

**Linux namespaces · cgroups · Bridge networking · REST API**  
Go implementation of a lightweight container orchestrator (mini-K8s)

[![Go](https://img.shields.io/badge/Go-1.22-blue?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

</div>

---

## Overview

Mini Orchestrator is a container orchestration system built from scratch in Go. It launches isolated containers using Linux namespaces (PID, NET, UTS, IPC, MOUNT), limits resources with cgroups, assigns IPs via a bridge network with NAT, and exposes a REST API for container lifecycle management. Designed as a senior-level portfolio project to demonstrate deep systems knowledge.

---

## Features

- **Container isolation** – unshare + pivot_root, separate network/pid/uts namespaces.
- **Resource limits** – CPU shares (cfs quota) and memory limits via cgroups.
- **Bridge networking** – veth pairs, IPAM, iptables MASQUERADE for outbound internet.
- **Scheduler** – binpacking (single node demo; extensible to multi‑node).
- **REST API** – create, list, get, delete containers.
- **Persistence** – container state in memory; rootfs on disk.

---

## Quick Start

### Prerequisites
- Linux host (Ubuntu/Debian) with root access.
- `iptables`, `ip`, `unshare`, `nsenter`, `pivot_root` (standard).

### Setup bridge (as root)
```bash
sudo ./scripts/setup-bridge.sh
Run orchestrator
bash
make run   # requires sudo
API usage
bash
# Create container
curl -X POST http://localhost:8080/containers \
  -H "Content-Type: application/json" \
  -d '{"image":"busybox","cmd":["/bin/sh","-c","sleep 300"],"cpu":500,"memory":134217728}'

# List containers
curl http://localhost:8080/containers

# Get container
curl http://localhost:8080/containers/<id>

# Delete
curl -X DELETE http://localhost:8080/containers/<id>
Architecture
text
REST API (gorilla/mux)
        │
        ▼
Container Manager
   ├── Process launcher (unshare + pivot_root)
   ├── Cgroup manager (CPU/memory)
   ├── Network manager (bridge + veth + IPAM)
   └── Scheduler (binpacking)
Notes
Requires root due to cgroups and network manipulation.

Rootfs is minimal; you can replace with real container images (e.g., extract Docker tar).

No image registry; you must pre-populate /var/lib/mini-containers/<id>/rootfs.

License
MIT.

</div> ```
14. Dockerfile (para construir el binario, no para ejecutar el orchestrator)
dockerfile
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o orchestrator ./cmd/orchestrator

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y iptables iproute2 util-linux && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/orchestrator /usr/local/bin/
CMD ["orchestrator"]