package types

type Container struct {
    ID     string `json:"id"`
    Image  string `json:"image"`
    Cmd    []string `json:"cmd"`
    CPU    int    `json:"cpu"`    // milicores (1000 = 1 core)
    Memory int64  `json:"memory"` // bytes
    Status string `json:"status"` // running, stopped, error
    IP     string `json:"ip"`
    Port   int    `json:"port"`   // puerto expuesto en host
}

type CreateContainerRequest struct {
    Image  string   `json:"image"`
    Cmd    []string `json:"cmd"`
    CPU    int      `json:"cpu"`
    Memory int64    `json:"memory"`
}

type CreateContainerResponse struct {
    ID string `json:"id"`
}

type ListContainersResponse struct {
    Containers []Container `json:"containers"`
}