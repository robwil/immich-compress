package immich

import (
	"fmt"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (c *ClientSimple) AssetDeleteMultiple(assetUUIDs []uuid.UUID, force bool) error {
	deleteDto := AssetBulkDeleteDto{
		Ids:   assetUUIDs,
		Force: &force,
	}

	resp, err := c.client.DeleteAssetsWithResponse(c.ctx, deleteDto)
	if err != nil {
		return fmt.Errorf("failed to delete assets: %w", err)
	}

	if resp.HTTPResponse.StatusCode >= 400 {
		return fmt.Errorf("bulk delete request failed with status %d: %s", resp.HTTPResponse.StatusCode, string(resp.Body))
	}

	return nil
}

func (c *ClientSimple) AssetDelete(assetID uuid.UUID, force bool) error {
	return c.AssetDeleteMultiple([]openapi_types.UUID{assetID}, force)
}
