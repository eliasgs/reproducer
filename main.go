// Reproducing issue https://github.com/fsouza/go-dockerclient/issues/374
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
)

func main() {
	var (
		c   *docker.Client
		err error
	)

	// create client
	if os.Getenv("DOCKER_HOST") != "" && os.Getenv("DOCKER_CERT_PATH") != "" {
		endpoint := os.Getenv("DOCKER_HOST")
		path := os.Getenv("DOCKER_CERT_PATH")
		ca := fmt.Sprintf("%s/ca.pem", path)
		cert := fmt.Sprintf("%s/cert.pem", path)
		key := fmt.Sprintf("%s/key.pem", path)
		c, err = docker.NewTLSClient(endpoint, cert, key, ca)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		endpoint := "unix:///var/run/docker.sock"
		c, err = docker.NewClient(endpoint)
		if err != nil {
			log.Fatal(err)
		}
	}

	// start a busysbox container
	container, err := c.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: "busybox",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = c.StartContainer(container.ID, nil)
	if err != nil {
		log.Fatal(err)
	}

	// create and start attached exec
	exec, err := c.CreateExec(docker.CreateExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Container:    container.ID,
		Cmd:          []string{"sh"},
	})
	err = dockerpty.StartExec(c, exec)
	if err != nil {
		log.Println(err)
	}

	// clean up
	err = c.RemoveContainer(docker.RemoveContainerOptions{
		ID:    container.ID,
		Force: true,
	})
	if err != nil {
		log.Fatal(err)
	}
}
