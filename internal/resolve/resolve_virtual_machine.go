package resolve

import (
	"context"
	"fmt"

	"github.com/magodo/aztft/internal/client"
	"github.com/magodo/aztft/internal/resourceid"
)

func resolveVirtualMachines(b *client.ClientBuilder, id resourceid.ResourceId) (string, error) {
	resourceGroupId := id.RootScope().(*resourceid.ResourceGroup)
	client, err := b.NewVirtualMachinesClient(resourceGroupId.SubscriptionId)
	if err != nil {
		return "", err
	}
	resp, err := client.Get(context.Background(), resourceGroupId.Name, id.Names()[0], nil)
	if err != nil {
		return "", fmt.Errorf("retrieving %q: %v", id, err)
	}
	props := resp.VirtualMachine.Properties
	if props == nil {
		return "", fmt.Errorf("unexpected nil property in response")
	}
	osProfile := props.OSProfile
	if osProfile == nil {
		return "", fmt.Errorf("unexpected nil OS profile in response")
	}

	switch {
	case osProfile.LinuxConfiguration != nil:
		return "azurerm_linux_virtual_machine", nil
	case osProfile.WindowsConfiguration != nil:
		return "azurerm_windows_virtual_machine", nil
	default:
		return "", fmt.Errorf("both windowsConfiguration and linuxConfiguration in OS profile is null")
	}
}
