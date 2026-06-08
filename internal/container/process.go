package container

import (
    "fmt"
    "os"
    "os/exec"
    "syscall"
    "path/filepath"
)

// Launch container with isolated namespaces (UTS, PID, NS, NET, IPC, USER)
func LaunchContainer(id, rootfs, hostname string, cmd []string) (int, error) {
    // Prepare command: pivot_root to rootfs and execute cmd
    // We'll use unshare + exec
    args := append([]string{"--fork", "--mount-proc", "--pid", "--net", "--uts", "--ipc"}, cmd...)
    c := exec.Command("unshare", args...)
    c.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
    }
    // Set root directory (chroot)
    c.Dir = rootfs
    c.SysProcAttr.Chroot = rootfs
    c.Stdin = os.Stdin
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr

    if err := c.Start(); err != nil {
        return 0, fmt.Errorf("start container: %v", err)
    }
    return c.Process.Pid, nil
}

// Setup cgroups for CPU and memory
func SetupCgroups(containerID string, cpuMillis int, memoryBytes int64) error {
    cgroupPath := filepath.Join("/sys/fs/cgroup", containerID)
    // Create cgroup directory
    if err := os.MkdirAll(cgroupPath, 0755); err != nil {
        return err
    }
    // Set CPU quota (period=100000us, quota=cpuMillis*1000)
    if cpuMillis > 0 {
        period := "100000"
        quota := fmt.Sprintf("%d", cpuMillis*1000)
        if err := os.WriteFile(filepath.Join(cgroupPath, "cpu.cfs_period_us"), []byte(period), 0644); err != nil {
            return err
        }
        if err := os.WriteFile(filepath.Join(cgroupPath, "cpu.cfs_quota_us"), []byte(quota), 0644); err != nil {
            return err
        }
    }
    // Set memory limit
    if memoryBytes > 0 {
        if err := os.WriteFile(filepath.Join(cgroupPath, "memory.limit_in_bytes"), []byte(fmt.Sprintf("%d", memoryBytes)), 0644); err != nil {
            return err
        }
    }
    return nil
}

// Assign a process to cgroup (write PID to cgroup.procs)
func AssignToCgroup(containerID string, pid int) error {
    cgroupPath := filepath.Join("/sys/fs/cgroup", containerID, "cgroup.procs")
    return os.WriteFile(cgroupPath, []byte(fmt.Sprintf("%d", pid)), 0644)
}