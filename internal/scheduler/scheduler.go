package scheduler

import "sync"

type Node struct {
    Name        string
    Containers  int
    CPUUsage    int
    MemoryUsage int64
}

type Scheduler struct {
    mu   sync.Mutex
    node *Node
}

func NewScheduler() *Scheduler {
    return &Scheduler{
        node: &Node{Name: "localhost", Containers: 0},
    }
}

func (s *Scheduler) Schedule(cpu int, mem int64) string {
    s.mu.Lock()
    defer s.mu.Unlock()
    // Siempre al localhost (único nodo)
    return s.node.Name
}

func (s *Scheduler) AddContainer() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.node.Containers++
}

func (s *Scheduler) RemoveContainer() {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.node.Containers > 0 {
        s.node.Containers--
    }
}