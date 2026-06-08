package api

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/tu-usuario/mini-orchestrator/internal/container"
    "github.com/tu-usuario/mini-orchestrator/pkg/types"
)

type Handler struct {
    cm *container.ContainerManager
}

func NewHandler(cm *container.ContainerManager) *Handler {
    return &Handler{cm: cm}
}

func (h *Handler) CreateContainer(w http.ResponseWriter, r *http.Request) {
    var req types.CreateContainerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    id, err := h.cm.CreateContainer(&req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(types.CreateContainerResponse{ID: id})
}

func (h *Handler) ListContainers(w http.ResponseWriter, r *http.Request) {
    containers := h.cm.ListContainers()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(types.ListContainersResponse{Containers: containers})
}

func (h *Handler) GetContainer(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    c, err := h.cm.GetContainer(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(c)
}

func (h *Handler) DeleteContainer(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if err := h.cm.DeleteContainer(id); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}