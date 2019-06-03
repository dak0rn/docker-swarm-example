package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	docker "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type DockerSwarmInfo struct {
	NodeID         string
	NodeAddr       string
	LocalNodeState string
	Error          string
}

type DockerHostInfo struct {
	ID                string
	Containers        int
	ContainersRunning int
	ContainersPaused  int
	ContainersStopped int
	Images            int
	SystemTime        string
	Name              string
	Swarm             DockerSwarmInfo
}

// Reads the given flag from the environment
// Panics, if it is unset
func readenv(name string) string {
	value, has := os.LookupEnv(name)
	if !has {
		panic(fmt.Sprintf("Environment variable %s not set", name))
	}

	return value
}

// Connects to the Docker host using parameter set in the environment
// and returns the connected client
// Will panic, if the connection fails.
func dockerConnect() *docker.Client {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if nil != err {
		panic(err)
	}

	return client
}

// Prints the given error and message to the console
func handleErr(err error, msg string) {
	log.WithField("error", err).Error(msg)
}

// Collects host information and sends it to the given
// master node
func collectAndSend(client *docker.Client, storeAddr string) {
	log.Info("Collecting host information...")

	info, err := client.Info(context.Background())
	if nil != err {
		handleErr(err, "Failed to retrieve host information")
		return
	}

	var serverData = DockerHostInfo{}
	serverData.Swarm = DockerSwarmInfo{}

	serverData.ID = info.ID
	serverData.Containers = info.Containers
	serverData.ContainersRunning = info.ContainersRunning
	serverData.ContainersPaused = info.ContainersPaused
	serverData.ContainersStopped = info.ContainersStopped
	serverData.Images = info.Images
	serverData.SystemTime = info.SystemTime
	serverData.Name = info.Name
	serverData.Swarm.NodeID = info.Swarm.NodeID
	serverData.Swarm.NodeAddr = info.Swarm.NodeAddr
	serverData.Swarm.LocalNodeState = string(info.Swarm.LocalNodeState)
	serverData.Swarm.Error = info.Swarm.Error

	jsonbytes, err := json.Marshal(serverData)
	if nil != err {
		handleErr(err, "Failed to encode the server info as JSON")
		return
	}

	log.Info("Sending data to the store...")
	_, err = http.Post(storeAddr, "application/json", bytes.NewReader(jsonbytes))
	if nil != err {
		handleErr(err, "Failed to send data to the store")
		return
	}

	log.Info("Information has been send.")
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: false,
		FullTimestamp:    true,
	})

	log.Info("Connecting to Docker...")
	client := dockerConnect()

	storeAddr := readenv("VIEWER_ADDR")

	log.Info("Starting the collection loop...")
	for {
		collectAndSend(client, storeAddr)
		<-time.After(10 * time.Second)
	}
}
