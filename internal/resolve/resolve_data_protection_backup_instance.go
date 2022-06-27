package resolve

import (
	"context"
	"fmt"
	"strings"

	"github.com/magodo/aztft/internal/client"
	"github.com/magodo/aztft/internal/resourceid"
)

func resolveDataProtectionBackupInstances(b *client.ClientBuilder, id resourceid.ResourceId) (string, error) {
	resourceGroupId := id.RootScope().(*resourceid.ResourceGroup)
	client, err := b.NewDataProtectionBackupInstancesClient(resourceGroupId.SubscriptionId)
	if err != nil {
		return "", err
	}
	resp, err := client.Get(context.Background(), resourceGroupId.Name, id.Names()[0], id.Names()[1], nil)
	if err != nil {
		return "", fmt.Errorf("retrieving %q: %v", id, err)
	}
	props := resp.BackupInstanceResource.Properties
	if props == nil {
		return "", fmt.Errorf("unexpected nil property in response")
	}
	dsinfo := props.DataSourceInfo
	if dsinfo == nil {
		return "", fmt.Errorf("unexpected nil properties.dataSourceInfo in response")
	}
	pdt := dsinfo.DatasourceType
	if pdt == nil {
		return "", fmt.Errorf("unexpected nil properties.dataSourceInfo.dataSourceType in response")
	}
	switch strings.ToUpper(*pdt) {
	case " MICROSOFT.DBFORPOSTGRESQL/SERVERS/DATABASES":
		return "azurerm_data_protection_backup_instance_postgresql", nil
	case "MICROSOFT.COMPUTE/DISKS":
		return "azurerm_data_protection_backup_instance_disk", nil
	case "MICROSOFT.STORAGE/STORAGEACCOUNTS/BLOBSERVICES":
		return "azurerm_data_protection_backup_instance_blob_storage", nil
	default:
		return "", fmt.Errorf("unknown data source type: %s", *pdt)
	}
}
