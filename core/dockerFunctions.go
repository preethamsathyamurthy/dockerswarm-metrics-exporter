package core

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	apiclient "github.com/docker/docker/client"
)

// Returns the Docker Root directory
func GetDockerRootInfo() (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return "", err
	}

	info, err := cli.Info(ctx)
	if err != nil {
		return "", err
	}

	return info.DockerRootDir, nil
}

// node inspect
// node list
// disk usage
// prune

func GetNodeList(hideTlsAndEngine bool) ([]swarm.Node, error) {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return []swarm.Node{}, err
	}

	list, err := cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return []swarm.Node{}, err
	}

	// Hiding the TLS information
	// Certificates are not necessary in metrics
	// Obtaining the TLS
	if hideTlsAndEngine {
		for index, _ := range list {
			list[index].Description.TLSInfo = swarm.TLSInfo{}
			list[index].Description.Engine = swarm.EngineDescription{}
		}
	}
	return list, err

}

// Reference returns the reference of a node. The special value "self" for a node
// reference is mapped to the current node, hence the node ID is retrieved using
// the `/info` endpoint.
// this function is private to this package
func getNodeference(ctx context.Context, client apiclient.APIClient, ref string) (string, error) {
	if ref == "self" {
		info, err := client.Info(ctx)
		if err != nil {
			return "", err
		}
		if info.Swarm.NodeID == "" {
			// If there's no node ID in /info, the node probably
			// isn't a manager. Call a swarm-specific endpoint to
			// get a more specific error message.
			_, err = client.NodeList(ctx, types.NodeListOptions{})
			if err != nil {
				return "", err
			}
			return "", errors.New("node ID not found in /info")
		}
		return info.Swarm.NodeID, nil
	}
	return ref, nil
}

// Inspects node with node ID
func GetNodeInspect(ref string) (swarm.Node, error) {

	//setting default ref to self
	if ref == "" {
		ref = "self"
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return swarm.Node{}, err
	}

	// Getting the ID of the node
	nodeRef, err := getNodeference(ctx, cli, ref)
	if err != nil {
		return swarm.Node{}, err
	}

	node, _, err := cli.NodeInspectWithRaw(ctx, nodeRef)
	if err != nil {
		return swarm.Node{}, err
	}
	return node, nil
}

// Is A Node Manager
func isANodeManager() (bool, *swarm.ManagerStatus, error) {
	node, err := GetNodeInspect("")
	if err != nil {
		return false, node.ManagerStatus, err
	}
	return true, node.ManagerStatus, err
}
