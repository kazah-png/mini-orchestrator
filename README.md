<div align="center">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=0:0d1117,50:001a2e,100:002b4d&height=130&section=header&text=mini-orchestrator&fontSize=38&fontColor=e6edf3&animation=fadeIn&fontAlignY=55" />
</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-Linux-FCC624?style=flat&logo=linux&logoColor=black)]()
[![License](https://img.shields.io/badge/License-MIT-3fb950?style=flat)](LICENSE)

**Lightweight container orchestrator written from scratch in Go.**  
Linux namespaces · cgroups · Bridge networking · IPAM · REST API

</div>

---

## Overview

`mini-orchestrator` launches isolated containers using Linux kernel primitives directly — no Docker daemon, no containerd, no runc. It creates new PID, NET, UTS, IPC, and MOUNT namespaces via `unshare`, pivots the root filesystem with `pivot_root`, limits CPU and memory with cgroups v1, assigns IP addresses from a managed pool, and wires each container to a bridge network with NAT for outbound connectivity.

A REST API manages the full container lifecycle: create, list, inspect, and delete.

---

## Architecture

```
REST API (gorilla/mux) :8080
          │
          ▼
  Container Manager
  ├── Process launcher   (unshare + pivot_root + exec)
  ├── Cgroup manager     (cpu.cfs_quota_us + memory.limit_in_bytes)
  ├── Network manager    (bridge + veth pair + IPAM + iptables MASQUERADE)
  └── Scheduler          (binpacking on declared CPU/memory)
          │
          ▼
  /var/lib/mini-containers/<id>/rootfs   (container filesystem)
```

---

## Features

| Feature | Details |
|---|---|
| **Namespace isolation** | PID, NET, UTS, IPC, MOUNT — each container gets its own namespace set |
| **Root pivot** | `pivot_root` into container rootfs; old root unmounted |
| **cgroups** | CPU shares via `cpu.cfs_quota_us`; memory cap via `memory.limit_in_bytes` |
| **Bridge network** | `mini-br0` bridge; veth pair per container; IPAM from `10.100.0.0/24` |
| **NAT** | `iptables -t nat -A POSTROUTING -j MASQUERADE` for outbound internet from containers |
| **REST API** | Create, list, get, delete container operations |
| **Scheduler** | Binpacking — assigns containers to nodes with the most remaining capacity |

---

## Prerequisites

- Linux host (Ubuntu 22.04 / Debian 12 recommended)
- Root access
- Standard tools: `ip`, `iptables`, `unshare`, `nsenter`, `pivot_root`

---

## Quick Start

### 1. Set up the bridge (once)

```bash
sudo ./scripts/setup-bridge.sh
```

This creates `mini-br0` at `10.100.0.1/24` and enables IP forwarding.

### 2. Run the orchestrator

```bash
make run   # runs as root
```

---

## API

### Create a container

```bash
curl -X POST http://localhost:8080/containers \
  -H "Content-Type: application/json" \
  -d '{
    "image":  "busybox",
    "cmd":    ["/bin/sh", "-c", "sleep 300"],
    "cpu":    500,
    "memory": 134217728
  }'
```

| Field | Description |
|---|---|
| `image` | Rootfs directory name under `/var/lib/mini-containers/images/` |
| `cmd` | Command to run inside the container |
| `cpu` | CPU quota in millicores (500 = 0.5 CPU) |
| `memory` | Memory limit in bytes |

### List containers

```bash
curl http://localhost:8080/containers
```

### Inspect a container

```bash
curl http://localhost:8080/containers/<id>
```

### Delete a container

```bash
curl -X DELETE http://localhost:8080/containers/<id>
```

---

## Preparing a rootfs

The orchestrator does not pull images from a registry. Provide a minimal root filesystem manually:

```bash
# Extract a Docker image as rootfs
docker export $(docker create busybox) | tar -C /var/lib/mini-containers/images/busybox -xf -
```

Or use `debootstrap` for a Debian base:

```bash
sudo debootstrap --arch=amd64 bookworm /var/lib/mini-containers/images/debian
```

---

## Networking detail

On container creation:
1. A `veth` pair is created: one end stays on the host, one moves into the container's network namespace.
2. The host end is attached to `mini-br0`.
3. An IP from `10.100.0.0/24` is assigned to the container end.
4. The container's default route is set to `10.100.0.1` (the bridge).
5. `iptables MASQUERADE` on the host provides outbound NAT.

---

## Limitations

- **Root required** — cgroups and network manipulation require elevated privileges.
- **cgroups v1 only** — systemd-based distros that use cgroups v2 need minor adapter changes.
- **No image registry** — rootfs must be pre-populated manually.
- **Single host** — no multi-node scheduling; the binpacker is a local demo.
- **In-memory state** — container metadata does not survive an orchestrator restart.

---

<div align="center">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=0:002b4d,50:001a2e,100:0d1117&height=80&section=footer" />
</div>
