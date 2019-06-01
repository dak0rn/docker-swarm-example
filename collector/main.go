package main

import (
	"context"
	"encoding/json"
	docker "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"time"
)

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
func collectAndSend(client *docker.Client) {
	log.Info("Collecting host information...")

	info, err := client.Info(context.Background())
	if nil != err {
		handleErr(err, "Failed to retrieve host information")
		return
	}

	jsonbytes, err := json.Marshal(info)
	if nil != err {
		handleErr(err, "Failed to encode the server info as JSON")
		return
	}

	log.Info("Data", string(jsonbytes))

	// TODO Send information to the master node using HTTP(S)
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: false,
		FullTimestamp:    true,
	})

	log.Info("Connecting to Docker...")
	client := dockerConnect()

	// TODO Fetch the master node address from the environment

	log.Info("Starting the collection loop...")
	for {
		collectAndSend(client)
		<-time.After(10 * time.Second)
	}
}
