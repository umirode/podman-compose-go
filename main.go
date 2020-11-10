package main

import (
	"fmt"
	"log"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
)

func main() {
	project, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: []string{"docker-compose.yml"},
		},
	}, nil)


	fmt.Println(project.GetServiceConfig())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(project.Config())
}
