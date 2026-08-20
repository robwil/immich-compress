package immich

import (
	"fmt"
	"net/http"

	"github.com/oapi-codegen/runtime/types"
)

func (c *ClientSimple) AssetDownload(id types.UUID) (*http.Response, error) {
	r, err := c.clientRaw.DownloadAsset(c.ctx, id, nil)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		r.Body.Close()
		return nil, fmt.Errorf("download failed for asset %s: status %d", id.String(), r.StatusCode)
	}
	return r, nil
}
