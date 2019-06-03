package main

import (
	"encoding/json"
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
	ContainerRunning  int
	ContainersPaused  int
	ContainersStopped int
	Images            int
	SystemTime        string
	Name              string
	Swarm             DockerSwarmInfo
}

func (self DockerHostInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(self)
}

func (self DockerHostInfo) UnmarshalBinary(data []byte) error {
	err := json.Unmarshal(data, &self)
	if nil != err {
		return err
	}

	return nil
}
