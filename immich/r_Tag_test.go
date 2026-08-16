package immich

import (
	"context"
	"net/http"
	"testing"
)

type mockTagClient struct {
	ClientWithResponsesInterface
	getAllTagsResp   *GetAllTagsResponse
	getAllTagsErr    error
	createTagResp   *CreateTagResponse
	createTagErr     error
}

func (m *mockTagClient) GetAllTagsWithResponse(_ context.Context, _ ...RequestEditorFn) (*GetAllTagsResponse, error) {
	return m.getAllTagsResp, m.getAllTagsErr
}

func (m *mockTagClient) CreateTagWithResponse(_ context.Context, _ CreateTagJSONRequestBody, _ ...RequestEditorFn) (*CreateTagResponse, error) {
	return m.createTagResp, m.createTagErr
}

func newTestClient(mock *mockTagClient) *ClientSimple {
	return &ClientSimple{
		client: mock,
		ctx:    context.Background(),
	}
}

func TestTagFindCreate_FindsExistingTag(t *testing.T) {
	tags := []TagResponseDto{
		{Id: "550e8400-e29b-41d4-a716-446655440000", Name: "__immich-compress__"},
	}
	mock := &mockTagClient{
		getAllTagsResp: &GetAllTagsResponse{
			JSON200: &tags,
			HTTPResponse: &http.Response{StatusCode: 200},
		},
	}
	c := newTestClient(mock)

	uuid, dto, err := c.tagFindCreate("__immich-compress__", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid.String() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("unexpected UUID: %s", uuid.String())
	}
	if dto == nil || dto.Name != "__immich-compress__" {
		t.Error("expected tag DTO to be returned")
	}
}

func TestTagFindCreate_CreatesWhenNotFound(t *testing.T) {
	emptyTags := []TagResponseDto{}
	created := TagResponseDto{Id: "660e8400-e29b-41d4-a716-446655440000", Name: "__compressed__"}
	mock := &mockTagClient{
		getAllTagsResp: &GetAllTagsResponse{
			JSON200: &emptyTags,
			HTTPResponse: &http.Response{StatusCode: 200},
		},
		createTagResp: &CreateTagResponse{
			JSON201: &created,
			HTTPResponse: &http.Response{StatusCode: 201},
		},
	}
	c := newTestClient(mock)

	uuid, dto, err := c.tagFindCreate("__compressed__", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uuid.String() != "660e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("unexpected UUID: %s", uuid.String())
	}
	if dto == nil || dto.Name != "__compressed__" {
		t.Error("expected created tag DTO to be returned")
	}
}

func TestTagFindCreate_NilJSON200ReturnsError(t *testing.T) {
	mock := &mockTagClient{
		getAllTagsResp: &GetAllTagsResponse{
			JSON200: nil,
			HTTPResponse: &http.Response{StatusCode: 401},
		},
	}
	c := newTestClient(mock)

	_, _, err := c.tagFindCreate("__immich-compress__", nil)
	if err == nil {
		t.Fatal("expected error for nil JSON200")
	}
}

func TestTagFindCreate_NilJSON201ReturnsError(t *testing.T) {
	emptyTags := []TagResponseDto{}
	mock := &mockTagClient{
		getAllTagsResp: &GetAllTagsResponse{
			JSON200: &emptyTags,
			HTTPResponse: &http.Response{StatusCode: 200},
		},
		createTagResp: &CreateTagResponse{
			JSON201: nil,
			HTTPResponse: &http.Response{StatusCode: 500},
		},
	}
	c := newTestClient(mock)

	_, _, err := c.tagFindCreate("__compressed__", nil)
	if err == nil {
		t.Fatal("expected error for nil JSON201")
	}
}
