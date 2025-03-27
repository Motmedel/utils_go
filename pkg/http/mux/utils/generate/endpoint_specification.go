package generate

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"fmt"
	motmedelGzip "github.com/Motmedel/utils_go/pkg/encoding/gzip"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	muxResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/static_content"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func MakeStrongEtag(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("\"%x\"", h.Sum(nil))
}

func makeStaticContentHeaders(
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
		entries = append(entries, &muxResponse.HeaderEntry{Name: "Cache-Control", Value: cacheControl})
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
	return makeStaticContentHeaders(parameter.ContentType, parameter.CacheControl, etag, lastModified)
}

var ExtensionToParameter = map[string]*StaticContentParameter{
	".html":  {ContentType: "text/html", CacheControl: "no-cache", CandidateForCompression: true},
	".css":   {ContentType: "text/css", CandidateForCompression: true},
	".js":    {ContentType: "text/javascript", CandidateForCompression: true},
	".map":   {ContentType: "application/json", CandidateForCompression: true},
	".webp":  {ContentType: "image/webp"},
	".avif":  {ContentType: "image/avif"},
	".woff2": {ContentType: "font/woff2"},
	".txt":   {ContentType: "text/plain", CandidateForCompression: true},
	".xml":   {ContentType: "text/xml", CandidateForCompression: true},
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

	errGroup, cancelCtx := errgroup.WithContext(context.Background())

loop:
	for _, contentEncoding := range supportedContentEncodings {
		select {
		case <-cancelCtx.Done():
			break loop
		default:
			errGroup.Go(
				func() error {
					switch contentEncoding {
					case "gzip":
						gzipData, err := motmedelGzip.MakeGzipData(data)
						if err != nil {
							return fmt.Errorf("make gzip data: %w", err)
						}

						if len(gzipData) >= len(data) {
							return nil
						}

						etag := MakeStrongEtag(gzipData)

						headers := []*muxResponse.HeaderEntry{
							{Name: "Content-Encoding", Value: contentEncoding},
							{Name: "ETag", Value: etag},
						}

						for _, headerEntry := range staticContent.Headers {
							switch strings.ToLower(headerEntry.Name) {
							case "content-type", "cache-control", "last-modified":
								headers = append(
									headers,
									&muxResponse.HeaderEntry{Name: headerEntry.Value, Value: headerEntry.Value},
								)
							}
						}

						contentEncodingToDataLock.Lock()
						defer contentEncodingToDataLock.Unlock()
						contentEncodingToData[contentEncoding] = &static_content.StaticContentData{
							Data:         gzipData,
							Etag:         etag,
							LastModified: staticContent.LastModified,
							Headers:      staticContent.Headers,
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
) (*endpoint_specification.EndpointSpecification, error) {
	extension := strings.ToLower(filepath.Ext(path))
	resultPath := "/" + path

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
		resultPath = strings.TrimSuffix(resultPath, ".html")
	}

	etag := MakeStrongEtag(data)

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

	return &endpoint_specification.EndpointSpecification{
		Path:          resultPath,
		Method:        http.MethodGet,
		StaticContent: staticContent,
	}, nil
}

func EndpointSpecificationsFromDirectory(
	rootPath string,
	addContentEncodingData bool,
) ([]*endpoint_specification.EndpointSpecification, error) {
	if rootPath == "" {
		return nil, nil
	}

	if !strings.HasSuffix(rootPath, "/") {
		rootPath += "/"
	}

	if rootPath == "" {
		return nil, nil
	}

	if !strings.HasSuffix(rootPath, "/") {
		rootPath += "/"
	}

	var specifications []*endpoint_specification.EndpointSpecification
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

func EndpointSpecificationsFromZip(reader *zip.Reader, addContentEncodingData bool) ([]*endpoint_specification.EndpointSpecification, error) {
	if reader == nil {
		return nil, nil
	}

	var specifications []*endpoint_specification.EndpointSpecification
	var specificationsMutex sync.Mutex

	errGroup, cancelCtx := errgroup.WithContext(context.Background())

loop:
	for _, file := range reader.File {
		select {
		case <-cancelCtx.Done():
			break loop
		default:
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

					suggestedEndpointPath := "/" + file.Name
					lastModified := file.FileInfo().ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT"

					specification, err := EndpointSpecificationFromDataPath(
						suggestedEndpointPath,
						data,
						lastModified,
						addContentEncodingData,
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
	}

	if err := errGroup.Wait(); err != nil {
		return nil, fmt.Errorf("errgroup wait: %w", err)
	}

	return specifications, nil
}
