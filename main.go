package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dihedron/openstack-mcp-server/backend"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/pagination"
	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// --- MCP Resource Structures (simplified for demonstration) ---


// NetworkInfo represents a simplified view of a Network for listing.
type NetworkInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VolumeInfo represents a simplified view of a Volume for listing.
type VolumeInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Size   int    `json:"size_gb"`
}

func main() {
	log.Println("Starting OpenStack MCP Server...")

	clients, err := backend.New()
	if err != nil {
		log.Fatalf("Error initializing OpenStack clients: %v", err)
	}
	log.Println("Successfully authenticated with OpenStack.")

	// Create a new MCP server
	s := server.NewMCPServer(
		"OpenStack Resource Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true), // Assuming this enables resource exposure
		server.WithLogging(),
		server.WithRecovery(),
	)

	// --- Virtual Machine Tools ---
	listVMsTool := mcp.NewTool(
		"ListVMs",
		mcp.WithDescription("Lists all Virtual Machines."),
		mcp.WithNoInput(), // No input parameters for listing
		mcp.WithToolExecutor(func(ctx context.Context, inputs mcp.ToolInputs, _ mcp.ToolOutputter) (map[string]interface{}, error) {
	)

	getVMTool := mcp.NewTool(
		"GetVMDetails",
		mcp.WithDescription("Gets detailed information about a specific Virtual Machine."),
		mcp.WithString("vm_id", mcp.Required(), mcp.WithDescription("The ID of the Virtual Machine.")),
		mcp.WithToolExecutor(func(ctx context.Context, inputs mcp.ToolInputs, _ mcp.ToolOutputter) (map[string]interface{}, error) {
			vmID, ok := inputs.GetString("vm_id")
			if !ok {
				return nil, fmt.Errorf("vm_id is required")
			}
			log.Printf("Executing GetVMDetails tool for VM ID: %s", vmID)
			server, err := servers.Get(context.TODO(), clients.ComputeClient, vmID).Extract()
			if err != nil {
				return nil, fmt.Errorf("error getting VM details for ID %s: %v", vmID, err)
			}
			return map[string]interface{}{"vm_details": server}, nil // Return the full gophercloud server struct
		}),
	)

	// --- Network Tools ---
	listNetworksTool := mcp.NewTool(
		"ListNetworks",
		mcp.WithDescription("Lists all Networks."),
		mcp.WithNoInput(),
		mcp.WithToolExecutor(func(ctx context.Context, inputs mcp.ToolInputs, _ mcp.ToolOutputter) (map[string]interface{}, error) {
			log.Println("Executing ListNetworks tool")
			var allNetworks []NetworkInfo

			pager := networks.List(clients.NetworkClient, networks.ListOpts{})
			err := pager.EachPage(context.TODO(), func(_ context.Context, page pagination.Page) (bool, error) {
				networkList, err := networks.ExtractNetworks(page)
				if err != nil {
					return false, fmt.Errorf("failed to extract networks: %v", err)
				}
				for _, net := range networkList {
					allNetworks = append(allNetworks, NetworkInfo{ID: net.ID, Name: net.Name})
				}
				return true, nil
			})
			if err != nil {
				return nil, fmt.Errorf("error listing networks: %v", err)
			}
			return map[string]interface{}{"networks": allNetworks}, nil
		}),
	)

	getNetworkTool := mcp.NewTool(
		"GetNetworkDetails",
		mcp.WithDescription("Gets detailed information about a specific Network."),
		mcp.WithString("network_id", mcp.Required(), mcp.WithDescription("The ID of the Network.")),
		mcp.WithToolExecutor(func(ctx context.Context, inputs mcp.ToolInputs, _ mcp.ToolOutputter) (map[string]interface{}, error) {
			networkID, ok := inputs.GetString("network_id")
			if !ok {
				return nil, fmt.Errorf("network_id is required")
			}
			log.Printf("Executing GetNetworkDetails tool for Network ID: %s", networkID)
			network, err := networks.Get(context.TODO(), clients.NetworkClient, networkID).Extract()
			if err != nil {
				return nil, fmt.Errorf("error getting network details for ID %s: %v", networkID, err)
			}
			return map[string]interface{}{"network_details": network}, nil
		}),
	)

	// --- Volume Tools ---
	listVolumesTool := mcp.NewTool(
		"ListVolumes",
		mcp.WithDescription("Lists all Volumes."),
		mcp.WithNoInput(),
		mcp.WithToolExecutor(func(ctx context.Context, inputs mcp.ToolInputs, _ mcp.ToolOutputter) (map[string]interface{}, error) {
			log.Println("Executing ListVolumes tool")
			var allVolumes []VolumeInfo

			pager := volumes.List(clients.VolumeClient, volumes.ListOpts{}) // Use appropriate ListOpts if needed
			err := pager.EachPage(context.TODO(), func(_ context.Context, page pagination.Page) (bool, error) {
				volumeList, err := volumes.ExtractVolumes(page)
				if err != nil {
					return false, fmt.Errorf("failed to extract volumes: %v", err)
				}
				for _, vol := range volumeList {
					allVolumes = append(allVolumes, VolumeInfo{
						ID:     vol.ID,
						Name:   vol.Name,
						Status: vol.Status,
						Size:   vol.Size,
					})
				}
				return true, nil
			})
			if err != nil {
				return nil, fmt.Errorf("error listing volumes: %v", err)
			}
			return map[string]interface{}{"volumes": allVolumes}, nil
		}),
	)

	getVolumeTool := mcp.NewTool(
		"GetVolumeDetails",
		mcp.WithDescription("Gets detailed information about a specific Volume."),
		mcp.WithString("volume_id", mcp.Required(), mcp.WithDescription("The ID of the Volume.")),
		mcp.WithToolExecutor(func(ctx context.Context, inputs mcp.ToolInputs, _ mcp.ToolOutputter) (map[string]interface{}, error) {
			volumeID, ok := inputs.GetString("volume_id")
			if !ok {
				return nil, fmt.Errorf("volume_id is required")
			}
			log.Printf("Executing GetVolumeDetails tool for Volume ID: %s", volumeID)
			volume, err := volumes.Get(context.TODO(), clients.VolumeClient, volumeID).Extract()
			if err != nil {
				return nil, fmt.Errorf("error getting volume details for ID %s: %v", volumeID, err)
			}
			return map[string]interface{}{"volume_details": volume}, nil
		}),
	)

	// Add tools to the server
	s.AddTools(
		listVMsTool, getVMTool,
		listNetworksTool, getNetworkTool,
		listVolumesTool, getVolumeTool,
	)

	log.Println("MCP Server setup complete. Starting to serve...")
	// Serve using stdio (typical for MCP tools) or an HTTP transport if mcp-go supports it.
	// The README example uses ServeStdio.
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
