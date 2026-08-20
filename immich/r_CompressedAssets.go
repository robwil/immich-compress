package immich

import (
	"fmt"

	"github.com/oapi-codegen/runtime/types"
)

// compressedAssetIDs returns asset IDs with the compressed tag for this client's user.
func (c *ClientSimple) compressedAssetIDs() (map[string]bool, error) {
	tagID := c.tags.compressedID
	if tagID == (types.UUID{}) {
		return map[string]bool{}, nil
	}

	ids := make(map[string]bool)
	var page float32 = 1
	tagIDs := []types.UUID{tagID}

	for {
		search := SearchAssetsJSONRequestBody{
			Page:   &page,
			TagIds: &tagIDs,
		}
		r, err := c.client.SearchAssetsWithResponse(c.ctx, search)
		if err != nil {
			return nil, fmt.Errorf("error searching compressed assets: %w", err)
		}
		if r.JSON200 == nil {
			return nil, fmt.Errorf("search compressed assets: status %d, body: %s", r.StatusCode(), string(r.Body))
		}

		for _, item := range r.JSON200.Assets.Items {
			ids[item.Id] = true
		}

		if r.JSON200.Assets.NextPage == nil {
			break
		}
		page++
	}

	return ids, nil
}

// CompressedAssetIDsAllUsers merges compressed asset IDs from the main client
// and all configured user key clients.
func (c *ClientSimple) CompressedAssetIDsAllUsers() (map[string]bool, error) {
	ids, err := c.compressedAssetIDs()
	if err != nil {
		return nil, err
	}

	if c.userKeys == nil {
		return ids, nil
	}

	for ownerID := range c.userKeys.Users {
		ownerClient, err := c.ClientForOwner(ownerID)
		if err != nil {
			return nil, err
		}
		if ownerClient == c {
			continue
		}
		if _, err := ownerClient.EnsureCompressedTag(); err != nil {
			return nil, fmt.Errorf("failed to resolve compressed tag for user %s: %w", ownerID, err)
		}

		ownerIDs, err := ownerClient.compressedAssetIDs()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch compressed assets for user %s: %w", ownerID, err)
		}
		for id := range ownerIDs {
			ids[id] = true
		}
	}

	return ids, nil
}
