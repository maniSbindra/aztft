package tfid

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/magodo/armid"
	"github.com/magodo/aztft/internal/client"
)

func buildStorageQueue(b *client.ClientBuilder, id armid.ResourceId, spec string) (string, error) {
	resourceGroupId := id.RootScope().(*armid.ResourceGroup)
	client, err := b.NewStorageAccountsClient(resourceGroupId.SubscriptionId)
	if err != nil {
		return "", err
	}
	resp, err := client.GetProperties(context.Background(), resourceGroupId.Name, id.Names()[0], nil)
	if err != nil {
		return "", fmt.Errorf("retrieving %q: %v", id, err)
	}
	props := resp.Account.Properties
	if props == nil {
		return "", fmt.Errorf("unexpected nil property in response")
	}
	endpoints := props.PrimaryEndpoints
	if endpoints == nil {
		return "", fmt.Errorf("unexpected nil properties.primaryEndpoints in response")
	}
	queueEndpoint := endpoints.Queue
	if queueEndpoint == nil {
		return "", fmt.Errorf("unexpected nil properties.primaryEndpoints.queue in response")
	}
	uri, err := url.Parse(*queueEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse url %s: %v", *queueEndpoint, err)
	}
	uri.Path = filepath.Join(uri.Path, id.Names()[2])
	return uri.String(), nil
}
