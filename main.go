package main

import (
    api "Capstone-4901---SeaTeam/api"
    config "Capstone-4901---SeaTeam/config"
    "Capstone-4901---SeaTeam/loadbalancer"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/fsnotify/fsnotify"
)

func main() {
    configuration, backendServers := loadConfig()
    lbPolicy := configuration.StaticResources.Clusters[0].LbPolicy

    var backendAddresses []string
    for _, server := range backendServers {
        backendAddresses = append(backendAddresses, fmt.Sprintf("http://%s:%d/", server.Address, server.Port))
    }

    var loadBalancer loadbalancer.LoadBalancer
    switch lbPolicy {
    case "ROUND_ROBIN":
        loadBalancer = loadbalancer.NewRoundRobinLoadBalancer(backendAddresses)
    case "LEAST_CONNECTIONS":
        loadBalancer = loadbalancer.NewLeastConnectionsLoadBalancer(backendAddresses)
    default:
        loadBalancer = loadbalancer.NewRoundRobinLoadBalancer(backendAddresses)
    }

    r := &Router{
        Timeout:      10 * time.Second,
        Config:       configuration,
        LoadBalancer: loadBalancer,
    }

    go watchConfigFile("config/static.yaml", func() {
        newConfig, newServers := loadConfig()
        r.Config = newConfig

        var updatedBackendAddresses []string
        for _, server := range newServers {
            updatedBackendAddresses = append(updatedBackendAddresses, fmt.Sprintf("http://%s:%d/", server.Address, server.Port))
        }

        switch newConfig.StaticResources.Clusters[0].LbPolicy {
        case "ROUND_ROBIN":
            r.LoadBalancer = loadbalancer.NewRoundRobinLoadBalancer(updatedBackendAddresses)
        case "LEAST_CONNECTIONS":
            r.LoadBalancer = loadbalancer.NewLeastConnectionsLoadBalancer(updatedBackendAddresses)
        default:
            r.LoadBalancer = loadbalancer.NewRoundRobinLoadBalancer(updatedBackendAddresses)
        }

        fmt.Println("Load balancer updated with new configuration")
    })

    http.Handle("/", r)
    http.HandleFunc("/health", api.HealthCheckHandler)
    http.HandleFunc("/endpoint1", api.Endpoint1Handler)
    http.HandleFunc("/endpoint2", api.Endpoint2Handler)

    fs := http.FileServer(http.Dir("./frontend"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    fmt.Println("Server started on :8000")
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
        fmt.Println("Error starting server:", err)
    }

    signalChannel := make(chan os.Signal, 1)
    signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
    <-signalChannel
}

func watchConfigFile(filePath string, reloadFunc func()) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        fmt.Println("Error creating file watcher:", err)
        return
    }
    defer watcher.Close()

    done := make(chan bool)
    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                if event.Op&fsnotify.Write == fsnotify.Write {
                    fmt.Println("Config file modified. Reloading...")
                    reloadFunc()
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                fmt.Println("Error watching config file:", err)
            }
        }
    }()

    err = watcher.Add(filePath)
    if err != nil {
        fmt.Println("Error watching config file:", err)
        return
    }

    <-done
}

func loadConfig() (config.StaticBootstrap, []config.BackendServer) {
    configuration, servers := config.GetYAMLdata()
    fmt.Println("Config reloaded successfully")
    return configuration, servers
}
