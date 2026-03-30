package cloud_storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/cloud_storage_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/bucket"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/object"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/object_list"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
)

const Domain = "storage.googleapis.com"

var defaultBaseUrl = &url.URL{
	Scheme: "https",
	Host:   Domain,
}

type Client struct {
	baseUrl       *url.URL
	uploadBaseUrl *url.URL
	config        *cloud_storage_config.Config
}

func NewClient(options ...cloud_storage_config.Option) *Client {
	return NewClientWithBaseUrl(defaultBaseUrl, options...)
}

func NewClientWithBaseUrl(baseUrl *url.URL, options ...cloud_storage_config.Option) *Client {
	u := *baseUrl
	u.Path = "/storage/v1/"

	uploadU := *baseUrl
	uploadU.Path = "/upload/storage/v1/"

	return &Client{baseUrl: &u, uploadBaseUrl: &uploadU, config: cloud_storage_config.New(options...)}
}

// InsertBucket creates a new bucket in the specified project.
func (c *Client) InsertBucket(ctx context.Context, project string, bucketConfig *bucket.Bucket, options ...fetch_config.Option) (*bucket.Bucket, error) {
	if project == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("project"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if bucketConfig == nil {
		return nil, nil
	}

	u := *c.baseUrl
	u.Path += "b"
	u.RawQuery = url.Values{"project": {project}}.Encode()
	urlString := u.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, createdBucket, err := motmedelHttpUtils.FetchJsonWithBody[*bucket.Bucket](ctx, urlString, bucketConfig, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdBucket, nil
}

// PatchBucket updates an existing bucket using patch semantics.
func (c *Client) PatchBucket(ctx context.Context, bucketName string, bucketConfig *bucket.Bucket, options ...fetch_config.Option) (*bucket.Bucket, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if bucketConfig == nil {
		return nil, nil
	}

	u := *c.baseUrl
	u.RawPath = u.Path + "b/" + url.PathEscape(bucketName)
	u.Path += "b/" + bucketName
	urlString := u.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPatch))
	_, patchedBucket, err := motmedelHttpUtils.FetchJsonWithBody[*bucket.Bucket](ctx, urlString, bucketConfig, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return patchedBucket, nil
}

// PatchObject updates an object's metadata using patch semantics.
func (c *Client) PatchObject(ctx context.Context, bucketName string, objectName string, objectConfig *object.Object, options ...fetch_config.Option) (*object.Object, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}
	if objectName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("object name"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if objectConfig == nil {
		return nil, nil
	}

	u := *c.baseUrl
	u.RawPath = u.Path + "b/" + url.PathEscape(bucketName) + "/o/" + url.PathEscape(objectName)
	u.Path += "b/" + bucketName + "/o/" + objectName
	urlString := u.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPatch))
	_, patchedObject, err := motmedelHttpUtils.FetchJsonWithBody[*object.Object](ctx, urlString, objectConfig, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return patchedObject, nil
}

// GetObject retrieves an object's metadata.
func (c *Client) GetObject(ctx context.Context, bucketName string, objectName string, options ...fetch_config.Option) (*object.Object, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}
	if objectName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("object name"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	u := *c.baseUrl
	u.RawPath = u.Path + "b/" + url.PathEscape(bucketName) + "/o/" + url.PathEscape(objectName)
	u.Path += "b/" + bucketName + "/o/" + objectName
	urlString := u.String()

	options = append(c.config.FetchOptions, options...)
	_, obj, err := motmedelHttpUtils.FetchJson[*object.Object](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return obj, nil
}

// DownloadObject downloads an object's content.
func (c *Client) DownloadObject(ctx context.Context, bucketName string, objectName string, options ...fetch_config.Option) ([]byte, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}
	if objectName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("object name"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	u := *c.baseUrl
	u.RawPath = u.Path + "b/" + url.PathEscape(bucketName) + "/o/" + url.PathEscape(objectName)
	u.Path += "b/" + bucketName + "/o/" + objectName
	u.RawQuery = url.Values{"alt": {"media"}}.Encode()
	urlString := u.String()

	options = append(c.config.FetchOptions, options...)
	_, responseBody, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return responseBody, nil
}

// ListObjects lists objects in a bucket. Use the query parameter to specify prefix, delimiter,
// maxResults, pageToken, and other query parameters.
func (c *Client) ListObjects(ctx context.Context, bucketName string, query url.Values, options ...fetch_config.Option) (*object_list.ObjectList, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	u := *c.baseUrl
	u.RawPath = u.Path + "b/" + url.PathEscape(bucketName) + "/o"
	u.Path += "b/" + bucketName + "/o"
	if query != nil {
		u.RawQuery = query.Encode()
	}
	urlString := u.String()

	options = append(c.config.FetchOptions, options...)
	_, list, err := motmedelHttpUtils.FetchJson[*object_list.ObjectList](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return list, nil
}

// DeleteObject deletes an object from a bucket.
func (c *Client) DeleteObject(ctx context.Context, bucketName string, objectName string, options ...fetch_config.Option) error {
	if bucketName == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}
	if objectName == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("object name"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	u := *c.baseUrl
	u.RawPath = u.Path + "b/" + url.PathEscape(bucketName) + "/o/" + url.PathEscape(objectName)
	u.Path += "b/" + bucketName + "/o/" + objectName
	urlString := u.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

// InsertObject uploads an object to a bucket using a multipart upload.
// The metadata should have at least its Name field set. The data parameter contains
// the object content, and contentType specifies its MIME type.
func (c *Client) InsertObject(ctx context.Context, bucketName string, metadata *object.Object, data []byte, contentType string, options ...fetch_config.Option) (*object.Object, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}
	if contentType == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("content type"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if metadata == nil {
		return nil, nil
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json marshal (metadata): %w", err))
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	metadataHeader := textproto.MIMEHeader{}
	metadataHeader.Set("Content-Type", "application/json; charset=UTF-8")
	metadataPart, err := writer.CreatePart(metadataHeader)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("create part (metadata): %w", err))
	}
	if _, err = metadataPart.Write(metadataBytes); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("write (metadata): %w", err))
	}

	mediaHeader := textproto.MIMEHeader{}
	mediaHeader.Set("Content-Type", contentType)
	mediaPart, err := writer.CreatePart(mediaHeader)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("create part (media): %w", err))
	}
	if _, err = mediaPart.Write(data); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("write (media): %w", err))
	}

	if err = writer.Close(); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("multipart writer close: %w", err))
	}

	u := *c.uploadBaseUrl
	u.RawPath = u.Path + "b/" + url.PathEscape(bucketName) + "/o"
	u.Path += "b/" + bucketName + "/o"
	u.RawQuery = url.Values{"uploadType": {"multipart"}}.Encode()
	urlString := u.String()

	options = append(
		append(c.config.FetchOptions, options...),
		fetch_config.WithMethod(http.MethodPost),
		fetch_config.WithBody(buf.Bytes()),
		fetch_config.WithHeaders(map[string]string{
			"Content-Type": "multipart/related; boundary=" + writer.Boundary(),
		}),
	)

	_, insertedObject, err := motmedelHttpUtils.FetchJson[*object.Object](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return insertedObject, nil
}
