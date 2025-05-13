package log

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/Motmedel/ecs_go/ecs"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelIter "github.com/Motmedel/utils_go/pkg/iter"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	motmedelTlsContext "github.com/Motmedel/utils_go/pkg/tls/context"
	motmedelTlsTypes "github.com/Motmedel/utils_go/pkg/tls/types"
	"log/slog"
	"strings"
)

const (
	timestampFormat = "2006-01-02T15:04:05.999999999Z"
)

func extractAlternativeNames(certificate *x509.Certificate) []string {
	if certificate == nil {
		return nil
	}

	var names []string

	names = append(names, certificate.DNSNames...)
	for _, ip := range certificate.IPAddresses {
		names = append(names, ip.String())
	}

	names = append(names, certificate.EmailAddresses...)
	for _, u := range certificate.URIs {
		names = append(names, u.String())
	}

	return motmedelIter.Set(names)
}

func EnrichWithConnectionState(base *ecs.Base, connectionState *tls.ConnectionState, client bool) {
	if base == nil {
		return
	}

	if connectionState == nil {
		return
	}

	ecsTls := base.Tls
	if ecsTls == nil {
		ecsTls = &ecs.Tls{}
		base.Tls = ecsTls
	}

	ecsTls.Cipher = tls.CipherSuiteName(connectionState.CipherSuite)
	ecsTls.Established = connectionState.HandshakeComplete
	ecsTls.NextProtocol = strings.ToLower(connectionState.NegotiatedProtocol)
	ecsTls.Resumed = connectionState.DidResume

	var protocolName string
	var protocolVersion string

	switch connectionState.Version {
	case tls.VersionSSL30:
		protocolName = "ssl"
		protocolVersion = "3"
	case tls.VersionTLS10:
		protocolName = "tls"
		protocolVersion = "1.0"
	case tls.VersionTLS11:
		protocolName = "tls"
		protocolVersion = "1.1"
	case tls.VersionTLS12:
		protocolName = "tls"
		protocolVersion = "1.2"
	case tls.VersionTLS13:
		protocolName = "tls"
		protocolVersion = "1.3"
	}

	if protocolName != "" || protocolVersion != "" {
		ecsTls.TlsProtocol = &ecs.TlsProtocol{Name: protocolName, Version: protocolVersion}
	}

	if serverName := connectionState.ServerName; serverName != "" {
		ecsTlsClient := ecsTls.Client
		if ecsTlsClient == nil {
			ecsTlsClient = &ecs.TlsClient{}
			ecsTls.Client = ecsTlsClient
		}

		ecsTlsClient.ServerName = serverName
	}

	if peerCertificates := connectionState.PeerCertificates; len(peerCertificates) > 0 {
		if leaf := peerCertificates[0]; leaf != nil {
			// TODO: Add more fields.

			issuer := leaf.Issuer.String()
			subject := leaf.Subject.String()
			notAfter := leaf.NotAfter.UTC().Format(timestampFormat)
			notBefore := leaf.NotBefore.UTC().Format(timestampFormat)

			if client {
				ecsTlsClient := ecsTls.Client
				if ecsTlsClient == nil {
					ecsTlsClient = &ecs.TlsClient{}
					ecsTls.Client = ecsTlsClient
				}

				ecsTlsClient.Issuer = issuer
				ecsTlsClient.Subject = subject
				ecsTlsClient.NotAfter = notAfter
				ecsTlsClient.NotBefore = notBefore
			} else {
				ecsTlsServer := ecsTls.Server
				if ecsTlsServer == nil {
					ecsTlsServer = &ecs.TlsServer{}
					ecsTls.Server = ecsTlsServer
				}

				ecsTlsServer.Issuer = issuer
				ecsTlsServer.Subject = subject
				ecsTlsServer.NotAfter = notAfter
				ecsTlsServer.NotBefore = notBefore
			}
		}
	}
}

func ParseTlsContext(tlsContext *motmedelTlsTypes.TlsContext) *ecs.Base {
	if tlsContext == nil {
		return nil
	}

	connectionState := tlsContext.ConnectionState
	if connectionState == nil {
		return nil
	}

	var base ecs.Base
	EnrichWithConnectionState(&base, connectionState, tlsContext.ClientSide)

	return &base
}

func ExtractTlsContext(ctx context.Context, record *slog.Record) error {
	if dnsContext, ok := ctx.Value(motmedelTlsContext.TlsContextKey).(*motmedelTlsTypes.TlsContext); ok && dnsContext != nil {
		base := ParseTlsContext(dnsContext)
		if base != nil {
			baseBytes, err := json.Marshal(base)
			if err != nil {
				return motmedelErrors.NewWithTrace(fmt.Errorf("json marshal (ecs base): %w", err), base)
			}

			var baseMap map[string]any
			if err = json.Unmarshal(baseBytes, &baseMap); err != nil {
				return motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal (ecs base map): %w", err), baseMap)
			}

			record.Add(motmedelLog.AttrsFromMap(baseMap)...)
		}
	}

	return nil
}

var TlsContextExtractor = motmedelLog.ContextExtractorFunction(ExtractTlsContext)
