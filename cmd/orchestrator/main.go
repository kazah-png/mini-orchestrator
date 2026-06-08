package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/tu-usuario/mini-orchestrator/internal/container"
    "github.com/tu-usuario/mini-orchestrator/internal/network"
    "github.com/tu-usuario/mini-orchestrator/internal/scheduler"
    "github.com/tu-usuario/mini-orchestrator/internal/api"
)

func main() {
    // Setup bridge and IPAM
    if err := network.SetupBridge(); err != nil {
        log.Fatalf("Failed to setup bridge: %v", err)
    }
    ipam := network.NewIPAM()
    sched := scheduler.NewScheduler()
    cm := container.NewContainerManager(ipam, sched)
    handler := api.NewHandler(cm)

    r := mux.NewRouter()
    r.HandleFunc("/containers", handler.CreateContainer).Methods("POST")
    r.HandleFunc("/containers", handler.ListContainers).Methods("GET")
    r.HandleFunc("/containers/{id}", handler.GetContainer).Methods("GET")
    r.HandleFunc("/containers/{id}", handler.DeleteContainer).Methods("DELETE")

    log.Println("Mini orchestrator listening on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}