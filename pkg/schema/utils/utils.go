package utils

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"slices"
	"strconv"
	"strings"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelNet "github.com/Motmedel/utils_go/pkg/net"
	"github.com/Motmedel/utils_go/pkg/net/types/domain_parts"
	"github.com/Motmedel/utils_go/pkg/net/types/flow_tuple"
	"github.com/Motmedel/utils_go/pkg/schema"
	motmedelTlsTypes "github.com/Motmedel/utils_go/pkg/tls/types"
	motmedelWhoisTypes "github.com/Motmedel/utils_go/pkg/whois/types"
)

const (
	timestampFormat = "2006-01-02T15:04:05.999999999Z"
)

const unknownPlaceholder = "(unknown)"

func MakeConnectionMessage(base *schema.Base, suffix string) string {
	if base == nil {
		return ""
	}

	var sourceIpAddress string
	var destinationIpAddress string
	var sourcePort int
	var destinationPort int

	if ecsSource := base.Source; ecsSource != nil {
		sourceIpAddress = ecsSource.Ip
		sourcePort = ecsSource.Port
	}

	if ecsDestination := base.Destination; ecsDestination != nil {
		destinationIpAddress = ecsDestination.Ip
		destinationPort = ecsDestination.Port
	}

	sourcePart := unknownPlaceholder
	if sourceIpAddress != "" && sourcePort != 0 {
		sourcePart = net.JoinHostPort(sourceIpAddress, strconv.Itoa(sourcePort))
	} else if sourceIpAddress != "" {
		sourcePart = sourceIpAddress
	} else if sourcePort != 0 {
		sourcePart = fmt.Sprintf(":%d", sourcePort)
	}

	destinationPart := unknownPlaceholder
	if destinationIpAddress != "" && destinationPort != 0 {
		destinationPart = net.JoinHostPort(destinationIpAddress, strconv.Itoa(destinationPort))
	} else if destinationIpAddress != "" {
		destinationPart = destinationIpAddress
	} else if destinationPort != 0 {
		destinationPart = fmt.Sprintf(":%d", destinationPort)
	}

	transportPart := unknownPlaceholder
	if ecsNetwork := base.Network; ecsNetwork != nil {
		if ecsNetwork.Transport != "" {
			transportPart = ecsNetwork.Transport
		} else if ecsNetwork.IanaNumber != "" {
			transportPart = fmt.Sprintf("(%s)", ecsNetwork.IanaNumber)
		}
	}

	var message string

	if sourcePart == unknownPlaceholder && destinationPart == unknownPlaceholder && transportPart == unknownPlaceholder {
		message = unknownPlaceholder
		if suffix != "" {
			message = suffix
		}
	} else {
		message = fmt.Sprintf("%s to %s %s", sourcePart, destinationPart, transportPart)
		if suffix != "" {
			message += fmt.Sprintf(" - %s", suffix)
		}
	}

	return message
}

func DefaultHeaderExtractorWithMasking(requestResponse any, maskNames []string, maskValue string) string {
	var header http.Header

	switch typedRequestResponse := requestResponse.(type) {
	case *http.Request:
		header = typedRequestResponse.Header
	case *http.Response:
		header = typedRequestResponse.Header
	default:
		return ""
	}

	var headerStrings []string

	for name, values := range header {
		shouldMask := slices.Contains(maskNames, strings.ToLower(name))
		for _, value := range values {
			if shouldMask {
				value = maskValue
			}
			headerStrings = append(headerStrings, fmt.Sprintf("%s: %s\r\n", name, value))
		}
	}

	return strings.Join(headerStrings, "")
}

func DefaultMaskedHeaderExtractor(requestResponse any) string {
	return DefaultHeaderExtractorWithMasking(
		requestResponse,
		[]string{"authorization", "cookie", "set-cookie"},
		"(MASKED)",
	)
}

func DefaultHeaderExtractor(requestResponse any) string {
	return DefaultHeaderExtractorWithMasking(requestResponse, nil, "")
}

func ParseHttp(
	request *http.Request,
	requestBodyData []byte,
	response *http.Response,
	responseBodyData []byte,
) (*schema.Base, error) {
	if request == nil && len(requestBodyData) == 0 && response == nil && len(responseBodyData) == 0 {
		return nil, nil
	}

	network := &schema.Network{Protocol: "http"}

	var source *schema.Target
	var destination *schema.Target
	var client *schema.Target
	var server *schema.Target

	var ecsUrl *schema.Url
	var userAgent *schema.UserAgent
	var httpVersion string

	var httpRequest *schema.HttpRequest
	if request != nil {
		requestUrl := request.URL
		originalUrl := requestUrl.String()

		hostSource := requestUrl.Host
		if hostSource == "" {
			hostSource = request.Host
		}
		hostUrl := &url.URL{Host: hostSource}
		trimmedHost := hostUrl.Hostname()

		host := trimmedHost
		if len(hostSource) > 0 && hostSource[0] == '[' {
			host = "[" + trimmedHost + "]"
		}

		domainParts := domain_parts.New(trimmedHost)

		requestHeader := request.Header

		var forwardedString string
		var xForwardedFor string
		if requestHeader != nil {
			forwardedString = requestHeader.Get("Forwarded")
			xForwardedFor = requestHeader.Get("X-Forwarded-For")
		}

		var port int
		if portString := requestUrl.Port(); portString != "" {
			port, _ = strconv.Atoi(portString)
		}

		if trimmedHost != "" || port != 0 {
			destination = &schema.Target{Address: trimmedHost, Port: port}
			if ip := net.ParseIP(trimmedHost); ip != nil {
				destination.Ip = trimmedHost
				if ipVersion := motmedelNet.GetIpVersion(&ip); ipVersion == 4 {
					network.Type = "ipv4"
				} else if ipVersion == 6 {
					network.Type = "ipv6"
				}
			} else {
				destination.Domain = trimmedHost
				if domainParts != nil {
					destination.RegisteredDomain = domainParts.RegisteredDomain
					destination.Subdomain = domainParts.Subdomain
					destination.TopLevelDomain = domainParts.TopLevelDomain
				}
			}
		}

		if destinationTcpAddr, ok := request.Context().Value(http.LocalAddrContextKey).(*net.TCPAddr); ok {
			if destination == nil {
				destination = &schema.Target{}
			}
			destination.Ip = destinationTcpAddr.IP.String()
			destination.Port = destinationTcpAddr.Port

			if ipVersion := motmedelNet.GetIpVersion(&destinationTcpAddr.IP); ipVersion == 4 {
				network.Type = "ipv4"
			} else if ipVersion == 6 {
				network.Type = "ipv6"
			}
		}

		var username string
		var password string
		if userInfo := requestUrl.User; userInfo != nil {
			username = userInfo.Username()
			password, _ = userInfo.Password()
		}

		// TODO: Maybe I can use `parseTarget()`?
		if remoteAddr := request.RemoteAddr; remoteAddr != "" {
			sourceIpAddress, sourcePort, err := motmedelNet.SplitAddress(remoteAddr)
			if err != nil {
				return nil, motmedelErrors.New(
					fmt.Errorf("split address: %w", err),
					remoteAddr,
				)
			}
			source = &schema.Target{Ip: sourceIpAddress, Port: sourcePort}
		}

		if userAgentOriginal := request.UserAgent(); userAgentOriginal != "" {
			userAgent = &schema.UserAgent{Original: userAgentOriginal}
		}

		ecsUrl = &schema.Url{
			Domain:   host,
			Fragment: requestUrl.Fragment,
			Original: originalUrl,
			Password: password,
			Path:     requestUrl.Path,
			Port:     port,
			Query:    requestUrl.RawQuery,
			Scheme:   requestUrl.Scheme,
			Username: username,
		}
		if domainParts != nil {
			ecsUrl.RegisteredDomain = domainParts.RegisteredDomain
			ecsUrl.Subdomain = domainParts.Subdomain
			ecsUrl.TopLevelDomain = domainParts.TopLevelDomain
		}

		var contentType string
		if requestHeader != nil {
			contentType = requestHeader.Get("Content-Type")
		}

		httpRequest = &schema.HttpRequest{
			ContentType: contentType,
			Method:      request.Method,
			Referrer:    request.Referer(),
		}

		httpVersionMajor := request.ProtoMajor
		httpVersionMinor := request.ProtoMinor

		if httpVersionMajor != 0 || httpVersionMinor != 0 {
			httpVersion = fmt.Sprintf("%d.%d", request.ProtoMajor, request.ProtoMinor)

			if strings.HasPrefix(httpVersion, "3.") {
				network.Transport = "udp"
				network.IanaNumber = "17"
			} else {
				network.Transport = "tcp"
				network.IanaNumber = "6"
			}

			if destination != nil && source != nil {
				destinationIp := net.ParseIP(destination.Ip)
				serverIp := net.ParseIP(source.Ip)
				destinationPort := destination.Port
				sourcePort := source.Port

				protocolNumber, _ := strconv.Atoi(network.IanaNumber)

				if destinationIp != nil && serverIp != nil && destinationPort != 0 && sourcePort != 0 && protocolNumber != 0 {
					flowTuple := flow_tuple.New(
						destinationIp,
						serverIp,
						uint16(destinationPort),
						uint16(sourcePort),
						uint8(protocolNumber),
					)
					if flowTuple != nil {
						if communityId := flowTuple.Hash(); communityId != "" {
							network.CommunityId = append(network.CommunityId, communityId)
						}
					}
				}
			}
		}

		if forwardedString == "" && xForwardedFor == "" {
			client = source
			server = destination
		} else {
			// TODO: Currently relies on `X-Forwarded-For` rather than `Forwarded`; using the latter
			//	entails the inclusion of an external parsing library, which is not acceptable.

			var serverIpAddress string

			forwardedForSplit := strings.Split(xForwardedFor, ",")
			if len(forwardedForSplit) > 0 {
				forwardedForIpAddress := strings.TrimSpace(forwardedForSplit[0])

				if ip := net.ParseIP(forwardedForIpAddress); ip != nil {
					client = &schema.Target{Ip: forwardedForIpAddress, Address: forwardedForIpAddress}
				}

				if len(forwardedForSplit) > 1 {
					serverIpAddressElement := forwardedForSplit[len(forwardedForSplit)-1]
					if ip := net.ParseIP(serverIpAddressElement); ip != nil {
						serverIpAddress = serverIpAddressElement
					}
				}
			}

			if destination != nil && destination.Domain != "" {
				destinationCopy := *destination
				server = &destinationCopy
				server.Ip = serverIpAddress
				server.Port = 0
				server.Address = server.Domain
			}
		}
	}

	if len(requestBodyData) != 0 {
		if httpRequest == nil {
			httpRequest = &schema.HttpRequest{}
		}
		httpRequest.Body = &schema.Body{Bytes: len(requestBodyData), Content: string(requestBodyData)}
		httpRequest.MimeType = http.DetectContentType(requestBodyData)
	}

	var httpResponse *schema.HttpResponse
	if response != nil {
		httpResponse = &schema.HttpResponse{
			StatusCode:  response.StatusCode,
			ContentType: response.Header.Get("Content-Type"),
		}
	}

	if len(responseBodyData) != 0 {
		if httpResponse == nil {
			httpResponse = &schema.HttpResponse{}
		}
		httpResponse.Body = &schema.Body{Bytes: len(responseBodyData), Content: string(responseBodyData)}
		httpResponse.MimeType = http.DetectContentType(responseBodyData)
	}

	var ecsHttp *schema.Http
	if httpRequest != nil || httpResponse != nil {
		ecsHttp = &schema.Http{Request: httpRequest, Response: httpResponse, Version: httpVersion}
	}

	if source == nil && ecsHttp == nil && destination == nil && ecsUrl == nil && userAgent == nil && network == nil {
		return nil, nil
	}

	return &schema.Base{
		Client:      client,
		Destination: destination,
		Http:        ecsHttp,
		Server:      server,
		Source:      source,
		Url:         ecsUrl,
		UserAgent:   userAgent,
		Network:     network,
	}, nil
}

func MakeHttpMessage(base *schema.Base) string {
	if base == nil {
		return ""
	}

	remoteAddress := "-"
	if source := base.Source; source != nil {
		if source.Ip != "" {
			remoteAddress = source.Ip
		}
	}

	userName := "-"
	if user := base.User; user != nil {
		if user.Name != "" {
			userName = user.Name
		} else if user.Email != "" {
			userName = user.Email
		}
	}

	requestLine := "-"
	referrer := "-"
	userAgentOriginal := "-"

	if ecsHttp := base.Http; ecsHttp != nil {
		if httpRequest := ecsHttp.Request; httpRequest != nil {
			method := httpRequest.Method
			if method == "" {
				method = "-"
			}

			path := "-"
			if ecsUrl := base.Url; ecsUrl != nil {
				if ecsUrl.Original != "" {
					path = ecsUrl.Original
				} else if ecsUrl.Path != "" {
					path = ecsUrl.Path
					if ecsUrl.Query != "" {
						path += "?" + ecsUrl.Query
					}
				}
			}

			proto := "-"
			if ecsHttp.Version != "" {
				proto = "HTTP/" + ecsHttp.Version
			}

			requestLine = fmt.Sprintf("%s %s %s", method, path, proto)

			if httpRequest.Referrer != "" {
				referrer = httpRequest.Referrer
			}
		}
	}

	if userAgent := base.UserAgent; userAgent != nil {
		if userAgent.Original != "" {
			userAgentOriginal = userAgent.Original
		}
	}

	statusCodeString := "-"
	bodyBytesString := "-"
	if ecsHttp := base.Http; ecsHttp != nil {
		if httpResponse := ecsHttp.Response; httpResponse != nil {
			if httpResponse.StatusCode != 0 {
				statusCodeString = strconv.Itoa(httpResponse.StatusCode)
			}
			if body := httpResponse.Body; body != nil {
				if body.Bytes != 0 {
					bodyBytesString = strconv.Itoa(body.Bytes)
				}
			}
		}
	}

	return fmt.Sprintf(
		"%s - %s \"%s\" %s %s \"%s\" \"%s\"",
		remoteAddress,
		userName,
		requestLine,
		statusCodeString,
		bodyBytesString,
		referrer,
		userAgentOriginal,
	)
}

func ParseHttpContext(httpContext *motmedelHttpTypes.HttpContext) (*schema.Base, error) {
	if httpContext == nil {
		return nil, nil
	}

	var user *schema.User
	if httpContextUser := httpContext.User; httpContextUser != nil {
		user = &schema.User{
			Domain:   httpContextUser.Domain,
			Email:    httpContextUser.Email,
			FullName: httpContextUser.FullName,
			Hash:     httpContextUser.Hash,
			Id:       httpContextUser.Id,
			Name:     httpContextUser.Name,
			Roles:    httpContextUser.Roles,
		}

		var group *schema.Group
		if httpContextUserGroup := httpContextUser.Group; httpContextUserGroup != nil {
			group = &schema.Group{
				Domain: httpContextUserGroup.Domain,
				Id:     httpContextUserGroup.Id,
				Name:   httpContextUserGroup.Name,
			}
		}
		user.Group = group
	}

	base, err := ParseHttp(
		httpContext.Request,
		httpContext.RequestBody,
		httpContext.Response,
		httpContext.ResponseBody,
	)
	if err != nil {
		return nil, motmedelErrors.New(
			fmt.Errorf("parse http: %w", err),
		)
	}

	if base != nil {
		base.User = user
		base.Message = MakeHttpMessage(base)
	}

	return base, nil
}

func parseTarget(rawAddress string, rawIpAddress string, rawPort int) (*schema.Target, error) {
	var target *schema.Target

	if rawIpAddress != "" {
		ipAddressUrl := fmt.Sprintf("fake://%s", rawIpAddress)
		urlParsedClientIpAddress, err := url.Parse(ipAddressUrl)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("url parse (crafted ip address url): %w", err),
				ipAddressUrl,
			)
		}

		port := rawPort

		if portString := urlParsedClientIpAddress.Port(); portString != "" {
			port, err = strconv.Atoi(portString)
			if err != nil {
				return nil, motmedelErrors.NewWithTrace(
					fmt.Errorf("strconv atoi (port string): %w", err),
					portString,
				)
			}
		}

		ipAddress := urlParsedClientIpAddress.Hostname()
		address := rawAddress
		if address != "" {
			address = ipAddress
		}

		target = &schema.Target{Address: address, Domain: rawAddress, Ip: ipAddress, Port: port}
	} else if rawAddress != "" {
		target = &schema.Target{
			Address: rawAddress,
			Domain:  rawAddress,
			Port:    rawPort,
		}
	}

	if target != nil {
		if domain := target.Domain; domain != "" {
			domainParts := domain_parts.New(domain)
			if domainParts != nil {
				target.RegisteredDomain = domainParts.RegisteredDomain
				target.Subdomain = domainParts.Subdomain
				target.TopLevelDomain = domainParts.TopLevelDomain
			}
		}
	}

	return target, nil
}

func ParseWhoisContext(whoisContext *motmedelWhoisTypes.WhoisContext) (*schema.Base, error) {
	if whoisContext == nil {
		return nil, nil
	}

	clientAddress := whoisContext.ClientAddress
	clientIpAddress := whoisContext.ClientIpAddress
	clientPort := whoisContext.ClientPort
	client, err := parseTarget(clientAddress, clientIpAddress, clientPort)
	if err != nil {
		return nil, motmedelErrors.New(
			fmt.Errorf("parse target (client data): %w", err),
			clientAddress, clientAddress, clientPort,
		)
	}
	var requestBody *schema.Body
	if requestData := whoisContext.RequestData; len(requestData) > 0 {
		requestBody = &schema.Body{Bytes: len(requestData), Content: string(requestData)}
	}

	serverAddress := whoisContext.ServerAddress
	serverIpAddress := whoisContext.ServerIpAddress
	serverPort := whoisContext.ServerPort
	server, err := parseTarget(serverAddress, serverIpAddress, serverPort)
	if err != nil {
		return nil, motmedelErrors.New(
			fmt.Errorf("parse target (server data): %w", err),
			serverAddress, serverIpAddress, serverPort,
		)
	}
	var responseBody *schema.Body
	if responseData := whoisContext.ResponseData; len(responseData) > 0 {
		responseBody = &schema.Body{Bytes: len(responseData), Content: string(responseData)}
	}

	var whois *schema.Whois
	if requestBody != nil || responseBody != nil {
		whois = &schema.Whois{}
		if requestBody != nil {
			whois.Request = &schema.WhoisRequest{Body: requestBody}
		}
		if responseBody != nil {
			whois.Response = &schema.WhoisResponse{Body: responseBody}
		}
	}

	return &schema.Base{
		Client:  client,
		Network: &schema.Network{Protocol: "whois", Transport: whoisContext.Transport},
		Server:  server,
		Whois:   whois,
	}, nil
}

func EventCreatedReplaceAttr(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return attr
	}

	switch attr.Key {
	case slog.TimeKey:
		return slog.Group("event", slog.Any("created", attr.Value))
	case slog.LevelKey:
		if value, ok := attr.Value.Any().(string); ok {
			return slog.Group("log", slog.String("level", strings.ToLower(value)))
		}
	case slog.MessageKey:
		attr.Key = "message"
	}

	return attr
}

func TimestampReplaceAttr(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return attr
	}

	switch attr.Key {
	case slog.TimeKey:
		attr.Key = "@timestamp"
	case slog.LevelKey:
		if value, ok := attr.Value.Any().(string); ok {
			return slog.Group("log", slog.String("level", strings.ToLower(value)))
		}
	case slog.MessageKey:
		attr.Key = "message"
	}

	return attr
}

func CommunityIdFromTargets(sourceTarget, destinationTarget *schema.Target, protocolNumber int) string {
	if sourceTarget == nil {
		return ""
	}

	if destinationTarget == nil {
		return ""
	}

	if protocolNumber == 0 {
		return ""
	}

	sourceTargetIp := net.ParseIP(sourceTarget.Ip)
	destinationTargetIp := net.ParseIP(destinationTarget.Ip)
	sourceTargetPort := sourceTarget.Port
	destinationTargetPort := destinationTarget.Port

	if sourceTargetIp == nil || destinationTargetIp == nil || sourceTargetPort == 0 || destinationTargetPort == 0 {
		return ""
	}

	flowTuple := flow_tuple.New(
		sourceTargetIp,
		destinationTargetIp,
		uint16(sourceTargetPort),
		uint16(destinationTargetPort),
		uint8(protocolNumber),
	)
	if flowTuple == nil {
		return ""
	}
	return flowTuple.Hash()
}

func EnrichWithTlsConnectionState(base *schema.Base, connectionState *tls.ConnectionState, clientInitiated bool) {
	if base == nil {
		return
	}

	if connectionState == nil {
		return
	}

	ecsTls := base.Tls
	if ecsTls == nil {
		ecsTls = &schema.Tls{}
		base.Tls = ecsTls
	}

	ecsTls.Cipher = tls.CipherSuiteName(connectionState.CipherSuite)
	ecsTls.Established = connectionState.HandshakeComplete
	ecsTls.NextProtocol = strings.ToLower(connectionState.NegotiatedProtocol)
	ecsTls.Resumed = connectionState.DidResume

	switch connectionState.Version {
	case tls.VersionSSL30:
		ecsTls.TlsProtocol = &schema.TlsProtocol{Name: "ssl", Version: "3"}
	case tls.VersionTLS10:
		ecsTls.TlsProtocol = &schema.TlsProtocol{Name: "tls", Version: "1.0"}
	case tls.VersionTLS11:
		ecsTls.TlsProtocol = &schema.TlsProtocol{Name: "tls", Version: "1.1"}
	case tls.VersionTLS12:
		ecsTls.TlsProtocol = &schema.TlsProtocol{Name: "tls", Version: "1.2"}
	case tls.VersionTLS13:
		ecsTls.TlsProtocol = &schema.TlsProtocol{Name: "tls", Version: "1.3"}
	}

	if serverName := connectionState.ServerName; serverName != "" {
		ecsTlsClient := ecsTls.Client
		if ecsTlsClient == nil {
			ecsTlsClient = &schema.TlsClient{}
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

			if !clientInitiated {
				ecsTlsClient := ecsTls.Client
				if ecsTlsClient == nil {
					ecsTlsClient = &schema.TlsClient{}
					ecsTls.Client = ecsTlsClient
				}

				ecsTlsClient.Issuer = issuer
				ecsTlsClient.Subject = subject
				ecsTlsClient.NotAfter = notAfter
				ecsTlsClient.NotBefore = notBefore
			} else {
				ecsTlsServer := ecsTls.Server
				if ecsTlsServer == nil {
					ecsTlsServer = &schema.TlsServer{}
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

func EnrichWithTlsContext(base *schema.Base, tlsContext *motmedelTlsTypes.TlsContext) {
	if base == nil {
		return
	}

	if tlsContext == nil {
		return
	}

	connectionState := tlsContext.ConnectionState
	if connectionState == nil {
		return
	}

	EnrichWithTlsConnectionState(base, connectionState, tlsContext.ClientInitiated)
}

func ParseEmailAddress(value string) (*schema.EmailAddress, error) {
	if value == "" {
		return nil, nil
	}

	addr, err := mail.ParseAddress(value)
	if err != nil {
		return nil, fmt.Errorf("mail parse address: %w", err)
	}
	return &schema.EmailAddress{Address: addr.Address, Name: addr.Name}, nil
}
