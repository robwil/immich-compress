package immich

import (
	"fmt"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (c *ClientSimple) deleteAssetsWithClient(client ClientWithResponsesInterface, assetUUIDs []uuid.UUID, force bool) error {
	deleteDto := AssetBulkDeleteDto{
		Ids:   assetUUIDs,
		Force: &force,
	}

	resp, err := client.DeleteAssetsWithResponse(c.ctx, deleteDto)
	if err != nil {
		return fmt.Errorf("failed to delete assets: %w", err)
	}

	if resp.HTTPResponse.StatusCode >= 400 {
		return fmt.Errorf("bulk delete request failed with status %d: %s", resp.HTTPResponse.StatusCode, string(resp.Body))
	}

	return nil
}

func (c *ClientSimple) AssetDeleteMultiple(assetUUIDs []uuid.UUID, force bool) error {
	return c.deleteAssetsWithClient(c.client, assetUUIDs, force)
}

func (c *ClientSimple) AssetDelete(assetID uuid.UUID, force bool) error {
	return c.AssetDeleteMultiple([]openapi_types.UUID{assetID}, force)
}

func (c *ClientSimple) AssetDeleteAsOwner(assetID uuid.UUID, ownerID string, force bool) error {
	apiKey, ok := c.userKeys.GetAPIKey(ownerID)
	if !ok {
		return c.AssetDelete(assetID, force)
	}

	ownerClient, err := newClientForAPIKey(c.baseURL, apiKey)
	if err != nil {
		return fmt.Errorf("failed to create client for owner %s: %w", ownerID, err)
	}

	return c.deleteAssetsWithClient(ownerClient, []openapi_types.UUID{assetID}, force)
}
