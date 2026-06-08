package container

import (
    "fmt"
    "os"
    "path/filepath"
)

type ResourceManager struct {
    baseCgroupPath string
}

func NewResourceManager() *ResourceManager {
    return &ResourceManager{baseCgroupPath: "/sys/fs/cgroup"}
}

func (rm *ResourceManager) CreateCgroup(containerID string, cpuMillis int, memoryBytes int64) error {
    cgPath := filepath.Join(rm.baseCgroupPath, containerID)
    if err := os.MkdirAll(cgPath, 0755); err != nil {
        return err
    }
    // CPU
    if cpuMillis > 0 {
        periodPath := filepath.Join(cgPath, "cpu.cfs_period_us")
        quotaPath := filepath.Join(cgPath, "cpu.cfs_quota_us")
        if err := os.WriteFile(periodPath, []byte("100000"), 0644); err != nil {
            return err
        }
        quota := cpuMillis * 1000
        if err := os.WriteFile(quotaPath, []byte(fmt.Sprintf("%d", quota)), 0644); err != nil {
            return err
        }
    }
    // Memory
    if memoryBytes > 0 {
        memPath := filepath.Join(cgPath, "memory.limit_in_bytes")
        if err := os.WriteFile(memPath, []byte(fmt.Sprintf("%d", memoryBytes)), 0644); err != nil {
            return err
        }
    }
    return nil
}

func (rm *ResourceManager) AddPid(containerID string, pid int) error {
    procsPath := filepath.Join(rm.baseCgroupPath, containerID, "cgroup.procs")
    return os.WriteFile(procsPath, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func (rm *ResourceManager) DeleteCgroup(containerID string) error {
    cgPath := filepath.Join(rm.baseCgroupPath, containerID)
    return os.RemoveAll(cgPath)
}