package cloud_storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/cloud_storage_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/bucket"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/object"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/object_list"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_storage/types/signer"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
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

// MaxSignedUrlExpires is the upper bound on V4 signed URL lifetime enforced by GCS.
const MaxSignedUrlExpires = 7 * 24 * time.Hour

// rfc3986Escape percent-encodes per RFC 3986 unreserved-character rules.
// Used for V4 canonical path segments and query parts: every byte outside
// A-Z / a-z / 0-9 / '-' / '_' / '.' / '~' is encoded as %XX (uppercase hex).
// Multi-byte UTF-8 is handled correctly because each byte is encoded individually.
func rfc3986Escape(s string) string {
	const upperhex = "0123456789ABCDEF"

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'):
			b.WriteByte(c)
		case c == '-' || c == '_' || c == '.' || c == '~':
			b.WriteByte(c)
		default:
			b.WriteByte('%')
			b.WriteByte(upperhex[c>>4])
			b.WriteByte(upperhex[c&0x0f])
		}
	}
	return b.String()
}

// rfc3986EscapeObjectPath escapes an object name preserving '/' literals as required
// by the V4 canonical request format.
func rfc3986EscapeObjectPath(name string) string {
	parts := strings.Split(name, "/")
	for i, p := range parts {
		parts[i] = rfc3986Escape(p)
	}
	return strings.Join(parts, "/")
}

// canonicalQueryString builds the canonical query string from sorted (key, value) pairs,
// percent-encoding both keys and values per RFC 3986. Sorting is by encoded key, which
// — for our standard X-Goog-* parameters — matches sorting by raw key.
func canonicalQueryString(values url.Values) string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		ek := rfc3986Escape(k)
		for _, v := range values[k] {
			parts = append(parts, ek+"="+rfc3986Escape(v))
		}
	}

	return strings.Join(parts, "&")
}

// SignedUrl returns a V4 signed URL for the given object. The URL grants temporary
// access to perform the given HTTP method against the object, valid for the given
// duration (capped at 7 days by GCS). The signer's identity is embedded in the URL —
// the runtime must be able to sign on behalf of that identity (e.g. an in-process
// RSA key, or via IAM Credentials' signBlob).
//
// Only the host header is signed; payloads are sent as UNSIGNED-PAYLOAD, so the
// resulting URL works with any request body content.
func (c *Client) SignedUrl(
	ctx context.Context,
	s signer.Signer,
	method string,
	bucketName string,
	objectName string,
	expires time.Duration,
) (string, error) {
	if s == nil {
		return "", motmedelErrors.NewWithTrace(nil_error.New("signer"))
	}
	if bucketName == "" {
		return "", motmedelErrors.NewWithTrace(empty_error.New("bucket name"))
	}
	if objectName == "" {
		return "", motmedelErrors.NewWithTrace(empty_error.New("object name"))
	}
	if method == "" {
		method = http.MethodGet
	}

	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context err: %w", err)
	}

	expiresSeconds := int64(expires / time.Second)
	if expiresSeconds <= 0 || expires > MaxSignedUrlExpires {
		return "", motmedelErrors.NewWithTrace(
			fmt.Errorf("expires out of range (must be > 0 and <= 7 days): %s", expires),
		)
	}

	signerEmail := s.Email()
	if signerEmail == "" {
		return "", motmedelErrors.NewWithTrace(empty_error.New("signer email"))
	}

	now := time.Now().UTC()
	datestamp := now.Format("20060102")
	datetime := now.Format("20060102T150405Z")

	credentialScope := datestamp + "/auto/storage/goog4_request"
	credential := signerEmail + "/" + credentialScope

	host := c.baseUrl.Host
	encodedPath := "/" + rfc3986Escape(bucketName) + "/" + rfc3986EscapeObjectPath(objectName)

	const signedHeaders = "host"
	canonicalHeaders := "host:" + host + "\n"

	queryValues := url.Values{}
	queryValues.Set("X-Goog-Algorithm", "GOOG4-RSA-SHA256")
	queryValues.Set("X-Goog-Credential", credential)
	queryValues.Set("X-Goog-Date", datetime)
	queryValues.Set("X-Goog-Expires", strconv.FormatInt(expiresSeconds, 10))
	queryValues.Set("X-Goog-SignedHeaders", signedHeaders)

	canonicalQuery := canonicalQueryString(queryValues)

	canonicalRequest := strings.Join([]string{
		method,
		encodedPath,
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		"UNSIGNED-PAYLOAD",
	}, "\n")

	canonicalRequestHash := sha256.Sum256([]byte(canonicalRequest))

	stringToSign := strings.Join([]string{
		"GOOG4-RSA-SHA256",
		datetime,
		credentialScope,
		hex.EncodeToString(canonicalRequestHash[:]),
	}, "\n")

	signature, err := s.Sign(ctx, []byte(stringToSign))
	if err != nil {
		return "", motmedelErrors.New(fmt.Errorf("signer sign: %w", err))
	}

	return c.baseUrl.Scheme + "://" + host + encodedPath + "?" + canonicalQuery +
		"&X-Goog-Signature=" + hex.EncodeToString(signature), nil
}
