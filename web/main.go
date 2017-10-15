package main

import (
    "fmt"

    "github.com/fsouza/go-dockerclient"
)

func main() {
    endpoint := "unix:///var/run/docker.sock"
    client, _ := docker.NewClient(endpoint)

    exposedCadvPort := map[docker.Port]struct{}{
        "8080/tcp": {}}

    createContConf := docker.Config{
        ExposedPorts: exposedCadvPort,
        Image:        "127.0.0.1:5000/gateway",
    }

    portBindings := map[docker.Port][]docker.PortBinding{
        "8080/tcp": {{HostIP: "0.0.0.0", HostPort: "8080"}},
    }

    createContHostConfig := docker.HostConfig{
        Binds:           []string{"/var/run:/var/run:rw", "/sys:/sys:ro", "/var/lib/docker:/var/lib/docker:ro"},
        PortBindings:    portBindings,
        PublishAllPorts: true,
        Privileged:      false,
    }

    createContOps := docker.CreateContainerOptions{
        Name:       "cadvisor",
        Config:     &createContConf,
        HostConfig: &createContHostConfig,
    }

    fmt.Printf("\nops = %s\n", createContOps)

    cont, err := client.CreateContainer(createContOps)
    if err != nil {
        fmt.Printf("create error = %s\n", err)
    }
    fmt.Printf("container = %s\n", cont)

    err = client.StartContainer(cont.ID, nil)
    if err != nil {
        fmt.Printf("start error = %s\n", err)
    }
    fmt.Printf("start = %s\n", err)
}
