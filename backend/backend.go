package backend

import (
	"context"
	"log"
	"os"

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
		panic(err)
	}

	providerClient, err := openstack.AuthenticatedClient(ctx, opts)
	if err != nil {
		panic(err)
	}

	region := os.Getenv("OS_REGION_NAME")
	if region == "" {
		log.Println("OS_REGION_NAME not set, using default region.")
		region = "RegionOne"
	}

	computeClient, err := openstack.NewComputeV2(providerClient, gophercloud.EndpointOpts{Region: region})
	if err != nil {
		panic(err)
	}

	networkClient, err := openstack.NewNetworkV2(providerClient, gophercloud.EndpointOpts{Region: region})
	if err != nil {
		panic(err)
	}

	volumeClient, err := openstack.NewBlockStorageV3(providerClient, gophercloud.EndpointOpts{Region: region})
	if err != nil {
		panic(err)
	}

	return &Client{
		ComputeClient:  computeClient,
		NetworkClient:  networkClient,
		VolumeClient:   volumeClient,
		ProviderClient: providerClient,
	}, nil
}
