package tfid

import (
	"fmt"
	"strings"

	"github.com/magodo/armid"
	"github.com/magodo/aztft/internal/client"
	"github.com/magodo/aztft/internal/resmap"
)

type builderFunc func(*client.ClientBuilder, armid.ResourceId, string) (string, error)

var dynamicBuilders = map[string]builderFunc{
	"azurerm_active_directory_domain_service":                                        buildActiveDirectoryDomainService,
	"azurerm_storage_object_replication":                                             buildStorageObjectReplication,
	"azurerm_storage_share":                                                          buildStorageShare,
	"azurerm_storage_container":                                                      buildStorageContainer,
	"azurerm_storage_queue":                                                          buildStorageQueue,
	"azurerm_storage_table":                                                          buildStorageTable,
	"azurerm_key_vault_key":                                                          buildKeyVaultKey,
	"azurerm_key_vault_secret":                                                       buildKeyVaultSecret,
	"azurerm_key_vault_certificate":                                                  buildKeyVaultCertificate,
	"azurerm_key_vault_certificate_issuer":                                           buildKeyVaultCertificateIssuer,
	"azurerm_key_vault_managed_storage_account":                                      buildKeyVaultStorageAccount,
	"azurerm_key_vault_managed_storage_account_sas_token_definition":                 buildKeyVaultStorageAccountSasTokenDefinition,
	"azurerm_storage_blob":                                                           buildStorageBlob,
	"azurerm_storage_share_directory":                                                buildStorageShareDirectory,
	"azurerm_storage_share_file":                                                     buildStorageShareFile,
	"azurerm_storage_table_entity":                                                   buildStorageTableEntity,
	"azurerm_storage_data_lake_gen2_filesystem":                                      buildStorageDfs,
	"azurerm_storage_data_lake_gen2_path":                                            buildStorageDfsPath,
	"azurerm_network_interface_security_group_association":                           buildNetworkInterfaceSecurityGroupAssociation,
	"azurerm_network_interface_application_gateway_backend_address_pool_association": buildNetworkInterfaceApplicationGatewayBackendAddressPoolAssociation,
	"azurerm_network_interface_application_security_group_association":               buildNetworkInterfaceApplicationSecurityGroupAssociation,
	"azurerm_virtual_desktop_workspace_application_group_association":                buildDesktopWorkspaceApplicationGroupAssociation,
}

func NeedsAPI(rt string) bool {
	_, ok := dynamicBuilders[rt]
	return ok
}

func DynamicBuild(id armid.ResourceId, rt string) (string, error) {
	id = id.Clone()

	importSpec, err := GetImportSpec(id, rt)
	if err != nil {
		return "", fmt.Errorf("getting import spec for %s as %s: %v", id, rt, err)
	}

	builder, ok := dynamicBuilders[rt]
	if !ok {
		return "", fmt.Errorf("unknown resource type: %q", rt)
	}

	b, err := client.NewClientBuilder()
	if err != nil {
		return "", fmt.Errorf("new API client builder: %v", err)
	}

	return builder(b, id, importSpec)
}

func StaticBuild(id armid.ResourceId, rt string) (string, error) {
	id = id.Clone()

	importSpec, err := GetImportSpec(id, rt)
	if err != nil {
		return "", fmt.Errorf("getting import spec for %s as %s: %v", id, rt, err)
	}

	rid, ok := id.(*armid.ScopedResourceId)
	if !ok {
		return id.String(), nil
	}

	switch rt {
	case "azurerm_app_service_slot_virtual_network_swift_connection":
		rid.AttrTypes[2] = "config"
	case "azurerm_iot_time_series_insights_access_policy":
		rid.AttrTypes[1] = "config"
	case "azurerm_synapse_workspace_sql_aad_admin":
		rid.AttrTypes[1] = "sqlAdministrators"
	case "azurerm_monitor_diagnostic_setting":
		// input: <target id>/providers/Microsoft.Insights/diagnosticSettings/setting1
		// tfid : <target id>|setting1
		id = id.ParentScope()
		return id.String() + "|" + rid.Names()[0], nil

	case "azurerm_synapse_role_assignment":
		pid := id.Parent()
		if err := pid.Normalize(importSpec); err != nil {
			return "", fmt.Errorf("normalizing id %q for %q with import spec %q: %v", pid.String(), rt, importSpec, err)
		}
		return pid.String() + "|" + id.Names()[1], nil
	case "azurerm_postgresql_active_directory_administrator":
		pid := id.Parent()
		if err := pid.Normalize(importSpec); err != nil {
			return "", fmt.Errorf("normalizing id %q for %q with import spec %q: %v", pid.String(), rt, importSpec, err)
		}
		return pid.String(), nil
	case "azurerm_servicebus_namespace_network_rule_set":
		pid := id.Parent()
		if err := pid.Normalize(importSpec); err != nil {
			return "", fmt.Errorf("normalizing id %q for %q with import spec %q: %v", pid.String(), rt, importSpec, err)
		}
		return pid.String(), nil
	}

	if importSpec != "" {
		if err := rid.Normalize(importSpec); err != nil {
			return "", fmt.Errorf("normalizing id %q for %q with import spec %q: %v", id.String(), rt, importSpec, err)
		}
	}
	return id.String(), nil
}

func GetImportSpec(id armid.ResourceId, rt string) (string, error) {
	resmap.Init()
	m := resmap.TF2ARMIdMap
	_ = m
	item, ok := resmap.TF2ARMIdMap[rt]
	if !ok {
		return "", fmt.Errorf("unknown resource type %q", rt)
	}

	if id.ParentScope() == nil {
		// For root scope resource id, the import spec is guaranteed to be only one.
		return item.ManagementPlane.ImportSpecs[0], nil
	}

	switch len(item.ManagementPlane.ImportSpecs) {
	case 0:
		// The ID is dynamically built (e.g. for property-like or some of the data plane only resources)
		return "", nil
	case 1:
		return item.ManagementPlane.ImportSpecs[0], nil
	default:
		// Needs to be matched with the scope. Or there might be zero import spec, as for the hypothetic resource ids.
		idscope := id.ParentScope().ScopeString()
		i := -1
		for idx, scope := range item.ManagementPlane.ParentScopes {
			if strings.EqualFold(scope, idscope) {
				i = idx
			}
		}
		if i == -1 {
			return "", fmt.Errorf("id %q doesn't correspond to resource type %q", id, rt)
		}
		return item.ManagementPlane.ImportSpecs[i], nil
	}
}
