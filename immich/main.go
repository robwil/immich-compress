// Package immich
package immich

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type ClientSimple struct {
	client    ClientWithResponsesInterface
	clientRaw ClientInterface
	ctx       context.Context
	parallel  int
	baseURL   string
	userKeys  *UserKeys
	tags      struct {
		compressedID types.UUID
	}
}

func normalizeBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(baseURL, "/api") {
		baseURL += "/api"
	}
	return baseURL
}

func newClientForAPIKey(baseURL string, apiKey string) (*ClientWithResponses, error) {
	return NewClientWithResponses(baseURL, WithRequestEditorFn(
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("x-api-key", apiKey)
			return nil
		}))
}

func NewClientSimple(ctx context.Context, parralel int, baseURL string, apiKey string, userKeys *UserKeys) (*ClientSimple, error) {
	baseURL = normalizeBaseURL(baseURL)

	client, err := newClientForAPIKey(baseURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	clientSimple := &ClientSimple{client: client, clientRaw: client.ClientInterface, ctx: ctx, parallel: parralel, baseURL: baseURL, userKeys: userKeys}

	tagCompressedAtID, err := clientSimple.tagCompressedAt()
	if err != nil {
		return nil, fmt.Errorf("can not get/create tags: %w", err)
	}

	clientSimple.tags = struct{ compressedID types.UUID }{compressedID: tagCompressedAtID}

	return clientSimple, nil
}

func (c *ClientSimple) ClientForOwner(ownerID string) (*ClientSimple, error) {
	apiKey, ok := c.userKeys.GetAPIKey(ownerID)
	if !ok {
		return c, nil
	}

	ownerClient, err := newClientForAPIKey(c.baseURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for owner %s: %w", ownerID, err)
	}

	return &ClientSimple{
		client:    ownerClient,
		clientRaw: ownerClient.ClientInterface,
		ctx:       c.ctx,
		parallel:  c.parallel,
		baseURL:   c.baseURL,
		userKeys:  c.userKeys,
		// tags left empty — EnsureCompressedTag will lazily create per-user
	}, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func UUUIDOfString(id string) (types.UUID, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		err = fmt.Errorf("failed to parse UUID '%s': %w", id, err)
	}
	return uuid, err
}
