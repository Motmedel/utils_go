package generate

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	motmedelGzip "github.com/Motmedel/utils_go/pkg/encoding/gzip"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint"
	muxResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/static_content"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	"golang.org/x/sync/errgroup"
)

// TODO: Put in the `endpoint_specification` type package?

const robotsTxtCacheControl = "public, max-age=86400"

func MakeStaticContentHeaders(
	contentType string,
	cacheControl string,
	etag string,
	lastModified string,
) []*muxResponse.HeaderEntry {
	var entries []*muxResponse.HeaderEntry

	if contentType != "" {
		entries = append(entries, &muxResponse.HeaderEntry{Name: "Content-Type", Value: contentType})

	}
	if etag != "" {
		entries = append(entries, &muxResponse.HeaderEntry{Name: "ETag", Value: etag})
	}

	if lastModified != "" {
		entries = append(entries, &muxResponse.HeaderEntry{Name: "Last-Modified", Value: lastModified})
	}

	if cacheControl != "" {
		entries = append(entries, &muxResponse.HeaderEntry{Name: "Cache-Control", Value: cacheControl, Overwrite: true})
	}

	return entries
}

var supportedContentEncodings = []string{"gzip"}

type StaticContentParameter struct {
	ContentType             string
	CacheControl            string
	CandidateForCompression bool
}

func (parameter *StaticContentParameter) HeaderEntries(etag string, lastModified string) []*muxResponse.HeaderEntry {
	return MakeStaticContentHeaders(parameter.ContentType, parameter.CacheControl, etag, lastModified)
}

var ExtensionToParameter = map[string]*StaticContentParameter{
	".html":  {ContentType: "text/html", CacheControl: "no-cache", CandidateForCompression: true},
	".css":   {ContentType: "text/css", CandidateForCompression: true},
	".js":    {ContentType: "text/javascript", CandidateForCompression: true},
	".mjs":   {ContentType: "text/javascript", CandidateForCompression: true},
	".map":   {ContentType: "application/json", CandidateForCompression: true},
	".svg":   {ContentType: "image/svg+xml", CandidateForCompression: true},
	".webp":  {ContentType: "image/webp"},
	".avif":  {ContentType: "image/avif"},
	".woff2": {ContentType: "font/woff2"},
	".txt":   {ContentType: "text/plain", CandidateForCompression: true},
	".xml":   {ContentType: "text/xml", CandidateForCompression: true},
	".pdf":   {ContentType: "application/pdf", CandidateForCompression: true},
}

func AddContentEncodingData(staticContent *static_content.StaticContent) error {
	if staticContent == nil {
		return nil
	}

	data := staticContent.Data
	if len(data) == 0 {
		return nil
	}

	contentEncodingToData := make(map[string]*static_content.StaticContentData)
	var contentEncodingToDataLock sync.Mutex

	errGroup, errGroupCtx := errgroup.WithContext(context.Background())

loop:
	for _, contentEncoding := range supportedContentEncodings {
		select {
		case <-errGroupCtx.Done():
			break loop
		default:
			errGroup.Go(
				func() error {
					switch contentEncoding {
					case "gzip":
						gzipData, err := motmedelGzip.MakeGzipData(context.Background(), data)
						if err != nil {
							return fmt.Errorf("make gzip data: %w", err)
						}

						if len(gzipData) >= len(data) {
							return nil
						}

						etag := motmedelHttpUtils.MakeStrongEtag(gzipData)

						headers := []*muxResponse.HeaderEntry{
							{Name: "Content-Encoding", Value: contentEncoding},
							{Name: "ETag", Value: etag},
						}

						for _, headerEntry := range staticContent.Headers {
							switch strings.ToLower(headerEntry.Name) {
							case "content-type", "cache-control", "last-modified":
								headers = append(
									headers,
									&muxResponse.HeaderEntry{
										Name:      headerEntry.Name,
										Value:     headerEntry.Value,
										Overwrite: headerEntry.Overwrite,
									},
								)
							}
						}

						contentEncodingToDataLock.Lock()
						defer contentEncodingToDataLock.Unlock()
						contentEncodingToData[contentEncoding] = &static_content.StaticContentData{
							Data:         gzipData,
							Etag:         etag,
							LastModified: staticContent.LastModified,
							Headers:      headers,
						}
					default:
						return motmedelErrors.NewWithTrace(
							fmt.Errorf("%w: %s", muxErrors.ErrUnexpectedContentEncoding, contentEncoding),
							contentEncoding,
						)
					}

					return nil
				},
			)
		}
	}

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("errgroup wait: %w", err)
	}

	if len(contentEncodingToData) != 0 {
		staticContent.ContentEncodingToData = contentEncodingToData
		staticContent.Headers = append(
			staticContent.Headers,
			&muxResponse.HeaderEntry{Name: "Vary", Value: "Accept-Encoding"},
		)
	}

	return nil
}

func EndpointSpecificationFromDataPath(
	path string,
	data []byte,
	lastModified string,
	addContentEncodingData bool,
	private bool,
) (*endpoint.Endpoint, error) {
	if path == "" {
		return nil, nil
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	extension := strings.ToLower(filepath.Ext(path))

	parameter, ok := ExtensionToParameter[extension]
	if !ok {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %s", muxErrors.ErrUnsupportedFileExtension, extension),
			extension,
		)
	}
	if parameter == nil {
		return nil, motmedelErrors.NewWithTrace(muxErrors.ErrNilHeaderParameter)
	}
	if parameter.ContentType == "" {
		return nil, motmedelErrors.NewWithTrace(muxErrors.ErrEmptyContentType)
	}

	if extension == ".html" {
		path = strings.TrimSuffix(path, ".html")
	}

	if path == "/index" {
		path = "/"
	}

	etag := motmedelHttpUtils.MakeStrongEtag(data)

	var visibility string
	if private {
		visibility = "private"
	} else {
		visibility = "public"
	}

	if parameter.CacheControl == "" {
		parameter.CacheControl = strings.Join(
			[]string{visibility, "max-age=31356000", "immutable"},
			", ",
		)
	}

	staticContent := &static_content.StaticContent{
		StaticContentData: static_content.StaticContentData{
			Data:         data,
			Etag:         etag,
			LastModified: lastModified,
			Headers:      parameter.HeaderEntries(etag, lastModified),
		},
	}

	if addContentEncodingData && parameter.CandidateForCompression && len(data) > 1000 {
		if err := AddContentEncodingData(staticContent); err != nil {
			return nil, motmedelErrors.New(
				fmt.Errorf("add content encoding data: %w", err),
				staticContent,
			)
		}
	}

	return &endpoint.Endpoint{
		Path:          path,
		Method:        http.MethodGet,
		StaticContent: staticContent,
	}, nil
}

func EndpointSpecificationsFromDirectory(
	rootPath string,
	addContentEncodingData bool,
	private bool,
) ([]*endpoint.Endpoint, error) {
	if rootPath == "" {
		return nil, nil
	}

	if !strings.HasSuffix(rootPath, "/") {
		rootPath += "/"
	}

	var specifications []*endpoint.Endpoint
	var specificationsMutex sync.Mutex

	errGroup, cancelCtx := errgroup.WithContext(context.Background())

	err := filepath.Walk(
		rootPath,
		func(path string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return motmedelErrors.NewWithTrace(fmt.Errorf("filepath walk func: %w", err), path)
			}

			if fileInfo.IsDir() {
				return nil
			}

			select {
			case <-cancelCtx.Done():
				return nil
			default:
				errGroup.Go(
					func() error {
						data, err := os.ReadFile(path)
						if err != nil {
							return motmedelErrors.NewWithTrace(fmt.Errorf("read file: %w", err), path)
						}

						suggestedEndpointPath := "/" + strings.TrimPrefix(path, rootPath)
						lastModified := fileInfo.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT"

						specification, err := EndpointSpecificationFromDataPath(
							suggestedEndpointPath,
							data,
							lastModified,
							addContentEncodingData,
							private,
						)
						if err != nil {
							return motmedelErrors.New(
								fmt.Errorf("endpoint specification from data path: %w", err),
								suggestedEndpointPath,
								data,
								lastModified,
							)
						}

						specificationsMutex.Lock()
						defer specificationsMutex.Unlock()
						specifications = append(specifications, specification)

						return nil
					},
				)
			}

			return nil
		},
	)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("filepath walk: %w", err), rootPath)
	}

	if err := errGroup.Wait(); err != nil {
		return nil, fmt.Errorf("errgroup wait: %w", err)
	}

	return specifications, nil
}

func EndpointSpecificationsFromZip(
	reader *zip.Reader,
	addContentEncodingData bool,
	private bool,
) ([]*endpoint.Endpoint, error) {
	if reader == nil {
		return nil, nil
	}

	var specifications []*endpoint.Endpoint
	var specificationsMutex sync.Mutex

	errGroup, cancelCtx := errgroup.WithContext(context.Background())

loop:
	for _, file := range reader.File {
		select {
		case <-cancelCtx.Done():
			break loop
		default:

			if file.FileInfo().IsDir() {
				continue
			}

			errGroup.Go(
				func() error {
					fileReader, err := file.Open()
					if err != nil {
						return motmedelErrors.NewWithTrace(fmt.Errorf("zip file open: %w", err), file)
					}

					data, err := io.ReadAll(fileReader)
					if err := fileReader.Close(); err != nil {
						return motmedelErrors.NewWithTrace(fmt.Errorf("zip file reader close: %w", err), fileReader)
					}
					if err != nil {
						return motmedelErrors.NewWithTrace(fmt.Errorf("io read all (zip file reader): %w", err), fileReader)
					}

					path := file.Name
					lastModified := file.FileInfo().ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT"

					specification, err := EndpointSpecificationFromDataPath(
						path,
						data,
						lastModified,
						addContentEncodingData,
						private,
					)
					if err != nil {
						return motmedelErrors.New(
							fmt.Errorf("endpoint specification from data path: %w", err),
							path,
							data,
							lastModified,
						)
					}

					specificationsMutex.Lock()
					defer specificationsMutex.Unlock()
					specifications = append(specifications, specification)

					return nil
				},
			)
		}
	}

	if err := errGroup.Wait(); err != nil {
		return nil, fmt.Errorf("errgroup wait: %w", err)
	}

	return specifications, nil
}

func MakeRobotsTxt(robotsTxt *motmedelHttpTypes.RobotsTxt) *endpoint.Endpoint {
	if robotsTxt == nil {
		return nil
	}

	robotsTxtString := robotsTxt.String()
	if robotsTxtString == "" {
		return nil
	}

	data := []byte(robotsTxtString)
	etag := motmedelHttpUtils.MakeStrongEtag(data)
	lastModified := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT"

	return &endpoint.Endpoint{
		Path:   "/robots.txt",
		Method: http.MethodGet,
		StaticContent: &static_content.StaticContent{
			StaticContentData: static_content.StaticContentData{
				Data:         data,
				Etag:         etag,
				LastModified: lastModified,
				Headers:      MakeStaticContentHeaders("text/plain", robotsTxtCacheControl, etag, lastModified),
			},
		},
	}
}
