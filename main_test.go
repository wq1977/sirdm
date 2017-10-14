package main

import (
	"log"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
)

func Test_Pull(t *testing.T) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Printf("client init fail when restart container:%+v", err)
		return
	}

	err = client.PullImage(docker.PullImageOptions{
		Repository: "127.0.0.1:5000/gateway",
	}, docker.AuthConfiguration{})
	if err != nil {
		log.Printf("%+v", err)
		return
	}
}
