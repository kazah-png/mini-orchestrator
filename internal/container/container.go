package container

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
    "github.com/tu-usuario/mini-orchestrator/internal/network"
    "github.com/tu-usuario/mini-orchestrator/pkg/types"
)

type ContainerManager struct {
    mu           sync.RWMutex
    containers   map[string]*types.Container
    ipam         *network.IPAM
    resourceMgr  *ResourceManager
    scheduler    *scheduler.Scheduler
}

func NewContainerManager(ipam *network.IPAM, sched *scheduler.Scheduler) *ContainerManager {
    return &ContainerManager{
        containers:  make(map[string]*types.Container),
        ipam:        ipam,
        resourceMgr: NewResourceManager(),
        scheduler:   sched,
    }
}

func (cm *ContainerManager) CreateContainer(req *types.CreateContainerRequest) (string, error) {
    id := generateID()
    ip, err := cm.ipam.Allocate()
    if err != nil {
        return "", err
    }
    // Simulate pulling image (just create rootfs dir)
    rootfs := filepath.Join("/var/lib/mini-containers", id, "rootfs")
    if err := os.MkdirAll(rootfs, 0755); err != nil {
        return "", err
    }
    // In real implementation, you'd unpack a container image tarball here.
    // For demo, we copy a minimal busybox static binary.
    copyBusybox(rootfs)

    // Prepare container object
    container := &types.Container{
        ID:     id,
        Image:  req.Image,
        Cmd:    req.Cmd,
        CPU:    req.CPU,
        Memory: req.Memory,
        Status: "creating",
        IP:     ip,
        Port:   0, // not implemented port mapping
    }

    // Launch process with namespaces
    cmd := req.Cmd // e.g., ["/bin/sh"]
    pid, err := LaunchContainer(id, rootfs, id[:12], cmd)
    if err != nil {
        return "", err
    }
    // Setup cgroups
    if err := cm.resourceMgr.CreateCgroup(id, req.CPU, req.Memory); err != nil {
        return "", err
    }
    if err := cm.resourceMgr.AddPid(id, pid); err != nil {
        return "", err
    }
    // Setup network: veth pair and assign IP
    if err := network.CreateVethPair(id, pid, ip); err != nil {
        return "", err
    }
    container.Status = "running"
    cm.mu.Lock()
    cm.containers[id] = container
    cm.mu.Unlock()
    cm.scheduler.AddContainer()

    // Monitor process and update status on exit
    go cm.waitForExit(id, pid)

    return id, nil
}

func (cm *ContainerManager) waitForExit(id string, pid int) {
    // Wait for process to finish
    process, _ := os.FindProcess(pid)
    state, _ := process.Wait()
    cm.mu.Lock()
    if cont, ok := cm.containers[id]; ok {
        if state.Exited() {
            cont.Status = "stopped"
        } else {
            cont.Status = "error"
        }
    }
    cm.mu.Unlock()
    cm.scheduler.RemoveContainer()
    // Cleanup network and cgroups
    cm.ipam.Release(cm.containers[id].IP)
    cm.resourceMgr.DeleteCgroup(id)
    // Also delete veth pair (omitted for brevity)
}

func (cm *ContainerManager) GetContainer(id string) (*types.Container, error) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    c, ok := cm.containers[id]
    if !ok {
        return nil, fmt.Errorf("container not found")
    }
    return c, nil
}

func (cm *ContainerManager) ListContainers() []types.Container {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    list := make([]types.Container, 0, len(cm.containers))
    for _, c := range cm.containers {
        list = append(list, *c)
    }
    return list
}

func (cm *ContainerManager) DeleteContainer(id string) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    c, exists := cm.containers[id]
    if !exists {
        return fmt.Errorf("container not found")
    }
    // Kill process
    // We'd store PID earlier; for simplicity, we can use "kill -9" on all processes in cgroup
    // but we'll just remove the container from map and clean resources.
    delete(cm.containers, id)
    cm.ipam.Release(c.IP)
    cm.resourceMgr.DeleteCgroup(id)
    return nil
}

func generateID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}

func copyBusybox(rootfs string) {
    // In real implementation, copy a busybox binary and libs.
    // For demo, we just create a dummy file.
    os.MkdirAll(filepath.Join(rootfs, "bin"), 0755)
    os.WriteFile(filepath.Join(rootfs, "bin/sh"), []byte("#!/bin/sh\necho hello\nsleep infinity"), 0755)
}