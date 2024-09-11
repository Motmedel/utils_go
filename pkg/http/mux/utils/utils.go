package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypes "github.com/Motmedel/utils_go/pkg/http/mux/types"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const staticCacheControl = "public,max-age=31356000,immutable"

var supportedContentEncodings = []string{"gzip", "deflate"}

func parseLastModifiedTimestamp(timestamp string) (time.Time, error) {
	if t, err := time.Parse(time.RFC1123, timestamp); err != nil {
		return time.Time{}, err
	} else {
		return t, nil
	}
}

func IfNoneMatchCacheHit(ifNoneMatchValue string, etag string) bool {
	if ifNoneMatchValue == "" || etag == "" {
		return false
	}

	return ifNoneMatchValue == etag
}

func IfModifiedSinceCacheHit(ifModifiedSinceValue string, lastModifiedValue string) (bool, error) {
	if ifModifiedSinceValue == "" || lastModifiedValue == "" {
		return false, nil
	}

	ifModifiedSinceTimestamp, err := parseLastModifiedTimestamp(ifModifiedSinceValue)
	if err != nil {
		return false, &muxErrors.BadIfModifiedSinceTimestamp{
			InputError: motmedelErrors.InputError{
				Message: "An error occurred when parsing a If-Modified-Since timestamp.",
				Cause:   err,
				Input:   ifModifiedSinceValue,
			},
		}
	}

	lastModifiedTimestamp, err := parseLastModifiedTimestamp(lastModifiedValue)
	if err != nil {
		return false, &motmedelErrors.InputError{
			Message: "An error occurred when parsing a Last-Modified timestamp.",
			Cause:   err,
			Input:   lastModifiedValue,
		}
	}

	return ifModifiedSinceTimestamp.Equal(lastModifiedTimestamp) || lastModifiedTimestamp.Before(ifModifiedSinceTimestamp), nil
}

func makeGzipData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := gzipWriter.Write(data); err != nil {
		return nil, &motmedelErrors.CauseError{
			Message: "An error occurred when writing data with a gzip writer.",
			Cause:   err,
		}
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, &motmedelErrors.CauseError{
			Message: "An error occurred when closing a gzip writer.",
			Cause:   err,
		}
	}

	return buf.Bytes(), nil
}

func makeDeflateData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	deflateWriter, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, &motmedelErrors.CauseError{
			Message: "An error occurred when creating a deflate writer.",
			Cause:   err,
		}
	}

	if _, err := deflateWriter.Write(data); err != nil {
		return nil, &motmedelErrors.CauseError{
			Message: "An error occurred when writing data with a deflate writer.",
			Cause:   err,
		}
	}
	if err := deflateWriter.Close(); err != nil {
		return nil, &motmedelErrors.CauseError{
			Message: "An error occurred when closing a deflate writer.",
			Cause:   err,
		}
	}

	return buf.Bytes(), nil
}

func makeStrongEtag(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func makeStaticContentHeaders(
	contentType string,
	contentEncoding string,
	etag string,
	lastModified string,
	cacheControl string,
) []*muxTypes.HeaderEntry {
	headerEntries := []*muxTypes.HeaderEntry{
		{
			Name:  "Content-Type",
			Value: contentType,
		},
	}

	if contentEncoding != "" {
		headerEntries = append(
			headerEntries,
			&muxTypes.HeaderEntry{
				Name:  "Content-Encoding",
				Value: contentEncoding,
			},
		)
	}

	headerEntries = append(
		headerEntries,
		[]*muxTypes.HeaderEntry{
			{
				Name:      "Cache-Control",
				Value:     cacheControl,
				Overwrite: true,
			},
			{
				Name:  "ETag",
				Value: etag,
			},
			{
				Name:  "Last-Modified",
				Value: lastModified,
			},
		}...,
	)

	if contentType != "text/html" {
		headerEntries = append(
			headerEntries,
			&muxTypes.HeaderEntry{
				Name:      "Content-Security-Policy",
				Value:     "default-src 'none'; frame-ancestors 'none'; base-uri 'none', form-action 'none'",
				Overwrite: true,
			},
		)
	}

	return headerEntries
}

func MakeHandlerSpecificationsFromDirectory(rootPath string) ([]*muxTypes.HandlerSpecification, error) {
	if rootPath == "" {
		return nil, nil
	}

	if !strings.HasSuffix(rootPath, "/") {
		rootPath += "/"
	}

	var handlerSpecifications []*muxTypes.HandlerSpecification
	var handlerSpecificationsMutex sync.Mutex
	var numGoroutines int

	errorChannel := make(chan struct {
		error
		string
	})

	err := filepath.Walk(
		rootPath,
		func(path string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return &motmedelErrors.InputError{
					Message: "An error occurred when about to process a walked file path.",
					Cause:   err,
					Input:   path,
				}
			}

			if fileInfo.IsDir() {
				return nil
			}

			numGoroutines += 1
			go func() {
				errorChannel <- struct {
					error
					string
				}{
					error: func() error {
						extension := strings.ToLower(filepath.Ext(path))

						var contentType string
						cacheControl := staticCacheControl

						relativePath, _ := strings.CutPrefix(path, rootPath)
						resultPath := "/" + relativePath

						switch extension {
						case ".html":
							contentType = "text/html"
							cacheControl = "no-cache"
							resultPath, _ = strings.CutSuffix(resultPath, ".html")
						case ".css":
							contentType = "text/css"
						case ".js":
							contentType = "text/javascript"
						case ".map":
							contentType = "application/json"
						case ".svg":
							contentType = "image/svg+xml"
						case ".avif":
							contentType = "image/avif"
						case ".woff2":
							contentType = "font/woff2"
						case ".txt":
							contentType = "text/plain"
						case ".xml":
							contentType = "text/xml"
						default:
							return &muxErrors.UnsupportedFileExtensionError{
								InputError: motmedelErrors.InputError{
									Message: "An unexpected file extension was encountered.",
									Input:   extension,
								},
							}
						}

						data, err := os.ReadFile(path)
						if err != nil {
							return &motmedelErrors.InputError{
								Message: "An error occurred when reading a file.",
								Cause:   err,
								Input:   path,
							}
						}

						etag := makeStrongEtag(data)
						lastModified := fileInfo.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT"

						staticContent := &muxTypes.StaticContent{
							StaticContentData: muxTypes.StaticContentData{
								Data:         data,
								Etag:         etag,
								LastModified: lastModified,
								Headers: makeStaticContentHeaders(
									contentType,
									"",
									etag,
									lastModified,
									cacheControl,
								),
							},
							ContentEncodingToData: make(map[string]*muxTypes.StaticContentData),
						}

						contentEncodingErrorChannel := make(chan struct {
							error
							string
						})
						var contentEncodingMapMutex sync.Mutex

						for _, contentEncoding := range supportedContentEncodings {
							contentEncoding := contentEncoding
							go func() {
								contentEncodingErrorChannel <- struct {
									error
									string
								}{
									error: func() error {
										switch contentEncoding {
										case "gzip":
											gzipData, err := makeGzipData(data)
											if err != nil {
												return &motmedelErrors.CauseError{
													Message: "An error occurred when making gzip data.",
													Cause:   err,
												}
											}

											gzipEtag := makeStrongEtag(data)

											contentEncodingMapMutex.Lock()
											staticContent.ContentEncodingToData[contentEncoding] = &muxTypes.StaticContentData{
												Data: gzipData,
												Etag: gzipEtag,
												Headers: makeStaticContentHeaders(
													contentType,
													contentEncoding,
													gzipEtag,
													lastModified,
													cacheControl,
												),
											}
											contentEncodingMapMutex.Unlock()
										case "deflate":
											deflateData, err := makeDeflateData(data)
											if err != nil {
												return &motmedelErrors.CauseError{
													Message: "An error occurred when making deflate data.",
													Cause:   err,
												}
											}

											deflateEtag := makeStrongEtag(data)

											contentEncodingMapMutex.Lock()
											staticContent.ContentEncodingToData[contentEncoding] = &muxTypes.StaticContentData{
												Data: deflateData,
												Etag: deflateEtag,
												Headers: makeStaticContentHeaders(
													contentType,
													contentEncoding,
													deflateEtag,
													lastModified,
													cacheControl,
												),
											}
											contentEncodingMapMutex.Unlock()
										default:
											return &muxErrors.UnexpectedContentEncodingError{
												InputError: motmedelErrors.InputError{
													Message: "An unexpected content encoding was encountered.",
													Input:   contentEncoding,
												},
											}
										}

										return nil
									}(),
									string: contentEncoding,
								}
							}()
						}

						for _, _ = range supportedContentEncodings {
							errData := <-contentEncodingErrorChannel
							if errData.error != nil {
								return &motmedelErrors.InputError{
									Message: "An error occurred when creating a encoded version of the content.",
									Cause:   errData.error,
									Input:   errData.string,
								}
							}
						}

						handlerSpecification := &muxTypes.HandlerSpecification{
							Path:          resultPath,
							Method:        "GET",
							StaticContent: staticContent,
						}

						handlerSpecificationsMutex.Lock()
						handlerSpecifications = append(handlerSpecifications, handlerSpecification)
						handlerSpecificationsMutex.Unlock()

						return nil
					}(),
					string: rootPath,
				}
			}()

			return nil
		},
	)
	if err != nil {
		return nil, &motmedelErrors.InputError{
			Message: "An error occurred when walking a file path.",
			Cause:   err,
			Input:   rootPath,
		}
	}

	for i := 0; i < numGoroutines; i++ {
		errData := <-errorChannel
		if errData.error != nil {
			return nil, &motmedelErrors.InputError{
				Message: "An error occurred when creating a handler specification for a file.",
				Cause:   errData.error,
				Input:   errData.string,
			}
		}
	}

	return handlerSpecifications, nil
}
