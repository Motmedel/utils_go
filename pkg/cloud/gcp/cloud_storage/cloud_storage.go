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

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/bucket"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/object"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/object_list"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
)

const Domain = "storage.googleapis.com"

var baseUrl = &url.URL{
	Scheme: "https",
	Host:   Domain,
	Path:   "/storage/v1/",
}

var uploadBaseUrl = &url.URL{
	Scheme: "https",
	Host:   Domain,
	Path:   "/upload/storage/v1/",
}

// InsertBucket creates a new bucket in the specified project.
func InsertBucket(ctx context.Context, project string, bucketConfig *bucket.Bucket, options ...fetch_config.Option) (*bucket.Bucket, error) {
	if project == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("project"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if bucketConfig == nil {
		return nil, nil
	}

	u := *baseUrl
	u.Path += "b"
	u.RawQuery = url.Values{"project": {project}}.Encode()
	urlString := u.String()

	options = append(options, fetch_config.WithMethod(http.MethodPost))
	_, createdBucket, err := motmedelHttpUtils.FetchJsonWithBody[*bucket.Bucket](ctx, urlString, bucketConfig, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdBucket, nil
}

// PatchBucket updates an existing bucket using patch semantics.
func PatchBucket(ctx context.Context, bucketName string, bucketConfig *bucket.Bucket, options ...fetch_config.Option) (*bucket.Bucket, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if bucketConfig == nil {
		return nil, nil
	}

	u := *baseUrl
	u.Path += "b/" + url.PathEscape(bucketName)
	urlString := u.String()

	options = append(options, fetch_config.WithMethod(http.MethodPatch))
	_, patchedBucket, err := motmedelHttpUtils.FetchJsonWithBody[*bucket.Bucket](ctx, urlString, bucketConfig, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return patchedBucket, nil
}

// GetObject retrieves an object's metadata.
func GetObject(ctx context.Context, bucketName string, objectName string, options ...fetch_config.Option) (*object.Object, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}
	if objectName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("object name"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	u := *baseUrl
	u.Path += "b/" + url.PathEscape(bucketName) + "/o/" + url.PathEscape(objectName)
	urlString := u.String()

	_, obj, err := motmedelHttpUtils.FetchJson[*object.Object](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return obj, nil
}

// DownloadObject downloads an object's content.
func DownloadObject(ctx context.Context, bucketName string, objectName string, options ...fetch_config.Option) ([]byte, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}
	if objectName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("object name"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	u := *baseUrl
	u.Path += "b/" + url.PathEscape(bucketName) + "/o/" + url.PathEscape(objectName)
	u.RawQuery = url.Values{"alt": {"media"}}.Encode()
	urlString := u.String()

	_, responseBody, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return responseBody, nil
}

// ListObjects lists objects in a bucket. Use the query parameter to specify prefix, delimiter,
// maxResults, pageToken, and other query parameters.
func ListObjects(ctx context.Context, bucketName string, query url.Values, options ...fetch_config.Option) (*object_list.ObjectList, error) {
	if bucketName == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	u := *baseUrl
	u.Path += "b/" + url.PathEscape(bucketName) + "/o"
	if query != nil {
		u.RawQuery = query.Encode()
	}
	urlString := u.String()

	_, list, err := motmedelHttpUtils.FetchJson[*object_list.ObjectList](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return list, nil
}

// InsertObject uploads an object to a bucket using a multipart upload.
// The metadata should have at least its Name field set. The data parameter contains
// the object content, and contentType specifies its MIME type.
func InsertObject(ctx context.Context, bucketName string, metadata *object.Object, data []byte, contentType string, options ...fetch_config.Option) (*object.Object, error) {
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

	u := *uploadBaseUrl
	u.Path += "b/" + url.PathEscape(bucketName) + "/o"
	u.RawQuery = url.Values{"uploadType": {"multipart"}}.Encode()
	urlString := u.String()

	options = append(
		options,
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
