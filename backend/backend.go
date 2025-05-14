package backend

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gophercloud/gophercloud/pagination"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
)

// Client holds the authenticated OpenStack service clients.
type Client struct {
	ComputeClient  *gophercloud.ServiceClient
	NetworkClient  *gophercloud.ServiceClient
	VolumeClient   *gophercloud.ServiceClient
	ProviderClient *gophercloud.ProviderClient
}

func New() (*Client, error) {

	ctx := context.Background()

	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		slog.Error("error creating OpenStack provider client", "error", err)
		return nil, err
	}

	providerClient, err := openstack.AuthenticatedClient(ctx, opts)
	if err != nil {
		slog.Error("error creating OpenStack provider client", "error", err)
		return nil, err
	}

	region := os.Getenv("OS_REGION_NAME")
	if region == "" {
		slog.Warn("OS_REGION_NAME not set, using default region.")
		region = "RegionOne"
	}

	computeClient, err := openstack.NewComputeV2(providerClient, gophercloud.EndpointOpts{Region: region})
	if err != nil {
		slog.Error("error creating Nova client", "error", err)
		return nil, err
	}

	networkClient, err := openstack.NewNetworkV2(providerClient, gophercloud.EndpointOpts{Region: region})
	if err != nil {
		slog.Error("error creating Neutron client", "error", err)
		return nil, err
	}

	volumeClient, err := openstack.NewBlockStorageV3(providerClient, gophercloud.EndpointOpts{Region: region})
	if err != nil {
		slog.Error("error creating Cinder client", "error", err)
		return nil, err
	}

	return &Client{
		ComputeClient:  computeClient,
		NetworkClient:  networkClient,
		VolumeClient:   volumeClient,
		ProviderClient: providerClient,
	}, nil
}

// VMInfo represents a simplified view of a Virtual Machine for listing.
type VMInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *Client) ListServers(ctx context.Context) (map[string]interface{}, error) {
	slog.Debug("listing VMs")
	var servers []VMInfo

	pager := servers.List(c.ComputeClient, servers.ListOpts{})
	err := pager.EachPage(ctx, func(ctx context.Context, page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			return false, fmt.Errorf("failed to extract servers: %v", err)
		}
		for _, srv := range serverList {
			servers = append(servers, VMInfo{ID: srv.ID, Name: srv.Name, Status: srv.Status})
		}
		return true, nil
	})
	if err != nil {
		slog.Error("error listing VMs", "error", err)
		return nil, fmt.Errorf("error listing VMs: %v", err)
	}
	return map[string]interface{}{"vms": servers}, nil
}
