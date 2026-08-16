package service

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	motmedelMux "github.com/Motmedel/utils_go/pkg/http/mux"
	endpointPkg "github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint/static_content"
	muxUtils "github.com/Motmedel/utils_go/pkg/http/mux/utils"
	"github.com/Motmedel/utils_go/pkg/http/service/service_config"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
	cspUtils "github.com/Motmedel/utils_go/pkg/http/utils/content_security_policy"
)

// staticContentEndpoint is a document the service serves as it is, of the given content type.
func staticContentEndpoint(path string, contentType string) *endpointPkg.Endpoint {
	data := []byte("content")

	return &endpointPkg.Endpoint{
		Path:   path,
		Method: http.MethodGet,
		Public: true,
		StaticContent: &static_content.StaticContent{
			StaticContentData: static_content.StaticContentData{
				Data: data,
				Headers: muxUtils.MakeStaticContentHeaders(
					contentType,
					"no-cache",
					"",
					"Mon, 02 Jan 2006 15:04:05 GMT",
				),
			},
		},
	}
}

// staticContentData is what an endpoint of the mux serves as it is, and fails the test where the
// endpoint is not one that does.
func staticContentData(t *testing.T, mux *motmedelMux.Mux, path string) string {
	t.Helper()

	endpoint := mux.Get(path, http.MethodGet)
	if endpoint == nil {
		t.Fatalf("no endpoint at %q", path)
	}

	staticContent := endpoint.StaticContent
	if staticContent == nil {
		t.Fatalf("the endpoint at %q serves no static content", path)
	}

	return string(staticContent.Data)
}

func TestNewWithUnknownProfile(t *testing.T) {
	t.Parallel()

	if _, err := New(service_config.WithProfile("something else")); err == nil {
		t.Error("new: got no error, want one")
	}
}

func TestNewWithProfile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                             string
		profile                          service_config.Profile
		expectedStrictTransportSecurity  bool
		expectedApiContentSecurityPolicy bool
		expectedRobotsTxt                bool
		expectedReporting                bool
	}{
		{
			// Nothing is said to crawlers: none of them gets past the authentication to reach what
			// a robots.txt would keep it from.
			name:                             "internal api",
			profile:                          service_config.ProfileInternalApi,
			expectedStrictTransportSecurity:  true,
			expectedApiContentSecurityPolicy: true,
		},
		{
			name:                             "public api",
			profile:                          service_config.ProfilePublicApi,
			expectedStrictTransportSecurity:  true,
			expectedApiContentSecurityPolicy: true,
			expectedRobotsTxt:                true,
		},
		{
			name:                            "internal web",
			profile:                         service_config.ProfileInternalWeb,
			expectedStrictTransportSecurity: true,
			expectedReporting:               true,
		},
		{
			name:                            "public web",
			profile:                         service_config.ProfilePublicWeb,
			expectedStrictTransportSecurity: true,
			expectedRobotsTxt:               true,
			expectedReporting:               true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service, err := New(
				service_config.WithHost("example.com"),
				service_config.WithProfile(testCase.profile),
			)
			if err != nil {
				t.Fatalf("new: %v", err)
			}

			mux := service.Mux

			_, hasStrictTransportSecurity := mux.DefaultHeaders[strictTransportSecurityHeaderName]
			if hasStrictTransportSecurity != testCase.expectedStrictTransportSecurity {
				t.Errorf(
					"strict transport security: got %t, want %t",
					hasStrictTransportSecurity,
					testCase.expectedStrictTransportSecurity,
				)
			}

			hasRobotsTxt := mux.Get("/robots.txt", http.MethodGet) != nil
			if hasRobotsTxt != testCase.expectedRobotsTxt {
				t.Errorf("robots.txt: got %t, want %t", hasRobotsTxt, testCase.expectedRobotsTxt)
			}

			// The policy a service that serves no documents answers every response with, and the
			// one a document is answered with, are one or the other rather than both.
			apiContentSecurityPolicy := mux.DefaultHeaders[contentSecurityPolicyHeaderName]
			if (apiContentSecurityPolicy != "") != testCase.expectedApiContentSecurityPolicy {
				t.Errorf(
					"api content security policy: got %q, want one: %t",
					apiContentSecurityPolicy,
					testCase.expectedApiContentSecurityPolicy,
				)
			}

			documentContentSecurityPolicy := mux.DefaultDocumentHeaders[contentSecurityPolicyHeaderName]
			if (documentContentSecurityPolicy != "") == testCase.expectedApiContentSecurityPolicy {
				t.Errorf(
					"document content security policy: got %q alongside the one for an api",
					documentContentSecurityPolicy,
				)
			}

			_, hasReportingEndpoints := mux.DefaultDocumentHeaders[reportingEndpointsHeaderName]
			if hasReportingEndpoints != testCase.expectedReporting {
				t.Errorf("reporting: got %t, want %t", hasReportingEndpoints, testCase.expectedReporting)
			}

			hasReportEndpoint := mux.Get(CspReportToPath, http.MethodPost) != nil
			if hasReportEndpoint != testCase.expectedReporting {
				t.Errorf("report endpoint: got %t, want %t", hasReportEndpoint, testCase.expectedReporting)
			}
		})
	}
}

// TestNewProfileIsOverriddenByLaterOptions verifies that what a profile decides is a starting point
// rather than a verdict.
func TestNewProfileIsOverriddenByLaterOptions(t *testing.T) {
	t.Parallel()

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithProfile(service_config.ProfilePublicWeb),
		service_config.WithReporting(false),
		service_config.WithRobotsTxt(false),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if _, found := service.Mux.DefaultDocumentHeaders[reportingEndpointsHeaderName]; found {
		t.Error("reporting was not turned off")
	}

	if service.Mux.Get("/robots.txt", http.MethodGet) != nil {
		t.Error("the robots.txt was not turned off")
	}

	// What was not overridden still holds.
	if _, found := service.Mux.DefaultHeaders[strictTransportSecurityHeaderName]; !found {
		t.Error("strict transport security was turned off along with the rest")
	}
}

// TestApiContentSecurityPolicy verifies that a service serving no documents still answers with a
// policy, and with exactly one: the policy for documents is dropped rather than left to be sent
// alongside it, a browser enforcing every policy it is sent.
func TestApiContentSecurityPolicy(t *testing.T) {
	t.Parallel()

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithApiContentSecurityPolicy(true),
		// The viewer styling is what the policy otherwise permits on top; see TestViewerStyleHashes.
		service_config.WithChromeXmlViewer(false),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	mux := service.Mux

	if got := mux.DefaultHeaders[contentSecurityPolicyHeaderName]; got != ApiContentSecurityPolicy {
		t.Errorf("content security policy: got %q, want %q", got, ApiContentSecurityPolicy)
	}

	if got, found := mux.DefaultDocumentHeaders[contentSecurityPolicyHeaderName]; found {
		t.Errorf("document content security policy: got %q, want none", got)
	}
}

// TestViewerStyleHashes verifies that the styles a browser's viewer applies are permitted by the
// policy the response actually carries -- which for a service that serves no documents is the one
// every response carries, that being where its problem details are answered under.
func TestViewerStyleHashes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		options        []service_config.Option
		expectedChrome bool
		expectedEdge   bool
	}{
		{
			name:           "a web service by default",
			options:        []service_config.Option{service_config.WithProfile(service_config.ProfilePublicWeb)},
			expectedChrome: true,
		},
		{
			name: "a web service serving pdf",
			options: []service_config.Option{
				service_config.WithProfile(service_config.ProfilePublicWeb),
				service_config.WithEdgePdfViewer(true),
			},
			expectedChrome: true,
			expectedEdge:   true,
		},
		{
			// A service that serves no documents still answers a 404 with a problem detail, which a
			// browser asks for as XML and renders through the same viewer.
			name:           "an api service by default",
			options:        []service_config.Option{service_config.WithProfile(service_config.ProfilePublicApi)},
			expectedChrome: true,
		},
		{
			name: "turned off",
			options: []service_config.Option{
				service_config.WithProfile(service_config.ProfilePublicWeb),
				service_config.WithChromeXmlViewer(false),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			options := append([]service_config.Option{service_config.WithHost("example.com")}, testCase.options...)

			service, err := New(options...)
			if err != nil {
				t.Fatalf("new: %v", err)
			}

			// The policy is read where the response carries it, which is what the service decided.
			policy := service.Mux.DefaultHeaders[contentSecurityPolicyHeaderName]
			if policy == "" {
				policy = service.Mux.DefaultDocumentHeaders[contentSecurityPolicyHeaderName]
			}

			for _, hash := range cspUtils.ChromeXmlViewerStyleHashes {
				if strings.Contains(policy, hash) != testCase.expectedChrome {
					t.Errorf("chrome xml viewer hash %q: want %t\n%s", hash, testCase.expectedChrome, policy)
				}
			}

			for _, hash := range cspUtils.EdgePdfViewerStyleHashes {
				if strings.Contains(policy, hash) != testCase.expectedEdge {
					t.Errorf("edge pdf viewer hash %q: want %t\n%s", hash, testCase.expectedEdge, policy)
				}
			}

			// A hash source matches the body of a style element as it is. Only Edge's viewer styles
			// through style attributes, which is what takes 'unsafe-hashes'; permitting it for
			// Chrome's would permit more than that viewer needs.
			if strings.Contains(policy, "'unsafe-hashes'") != testCase.expectedEdge {
				t.Errorf("unsafe-hashes: want %t\n%s", testCase.expectedEdge, policy)
			}

			if testCase.expectedChrome || testCase.expectedEdge {
				styleSrc, _, found := strings.Cut(policy[strings.Index(policy, "style-src "):], ";")
				if !found {
					styleSrc = policy[strings.Index(policy, "style-src "):]
				}

				if expected := "style-src 'self'"; !strings.HasPrefix(styleSrc, expected) {
					t.Errorf("style-src: got %q, want it to start with %q", styleSrc, expected)
				}
			}
		})
	}
}

func TestTrustedTypes(t *testing.T) {
	t.Parallel()

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithProfile(service_config.ProfilePublicWeb),
		service_config.WithTrustedTypes("default", "dompurify"),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	policy := service.Mux.DefaultDocumentHeaders[contentSecurityPolicyHeaderName]

	for _, expected := range []string{
		"trusted-types default dompurify",
		// Naming the policies is worth nothing without requiring them.
		"require-trusted-types-for 'script'",
	} {
		if !strings.Contains(policy, expected) {
			t.Errorf("the policy lacks %q:\n%s", expected, policy)
		}
	}
}

func TestFedCm(t *testing.T) {
	t.Parallel()

	provider := &url.URL{Scheme: "https", Host: "accounts.example.com"}

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithProfile(service_config.ProfilePublicWeb),
		service_config.WithFedCm(provider),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	mux := service.Mux

	// The browser fetches the provider's configuration and accounts over connect-src.
	policy := mux.DefaultDocumentHeaders[contentSecurityPolicyHeaderName]
	if !strings.Contains(policy, "accounts.example.com") {
		t.Errorf("the policy does not permit the provider:\n%s", policy)
	}

	// Without the permission the browser refuses the call whatever the policy permits.
	permissionsPolicy := mux.DefaultDocumentHeaders[permissionsPolicyHeaderName]
	if expected := `identity-credentials-get=(self "https://accounts.example.com")`; !strings.Contains(permissionsPolicy, expected) {
		t.Errorf("the permissions policy lacks %q:\n%s", expected, permissionsPolicy)
	}

	// What the permissions policy said before is kept.
	if !strings.Contains(permissionsPolicy, "geolocation=()") {
		t.Errorf("the permissions policy lost what it said before:\n%s", permissionsPolicy)
	}
}

// TestApiContentSecurityPolicyParses verifies that what the service answers with is a policy this
// module can read back, rather than a string only a browser ever looks at.
func TestApiContentSecurityPolicyParses(t *testing.T) {
	t.Parallel()

	csp, err := content_security_policy.Parse([]byte(ApiContentSecurityPolicy))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if csp == nil {
		t.Fatal("nil content security policy")
	}

	defaultSrc := csp.GetDefaultSrc()
	if defaultSrc == nil {
		t.Fatal("no default-src")
	}
	if len(defaultSrc.Sources) != 1 || defaultSrc.Sources[0].String() != "'none'" {
		t.Errorf("default-src: got %v, want 'none'", defaultSrc.Sources)
	}

	// The directives that do not fall back to default-src are said in full.
	for _, name := range []string{"base-uri", "form-action", "frame-ancestors", "sandbox"} {
		if _, found := csp.GetDirective(name); !found {
			t.Errorf("no %s", name)
		}
	}
}

func TestStrictTransportSecurity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		host     string
		expected bool
	}{
		{name: "a host of its own", host: "example.com", expected: true},
		{name: "no host", host: "", expected: true},
		// A browser told to reach localhost over HTTPS only cannot reach a development server.
		{name: "localhost", host: "localhost"},
		{name: "a subdomain of localhost", host: "service.localhost"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			options := []service_config.Option{service_config.WithStrictTransportSecurity(true)}
			if testCase.host != "" {
				options = append(options, service_config.WithHost(testCase.host))
			}

			service, err := New(options...)
			if err != nil {
				t.Fatalf("new: %v", err)
			}

			value := service.Mux.DefaultHeaders[strictTransportSecurityHeaderName]
			if (value != "") != testCase.expected {
				t.Errorf("strict transport security: got %q, want a header: %t", value, testCase.expected)
			}
		})
	}
}

func TestRobotsTxt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		options           []service_config.Option
		expectedToContain []string
		expectedToLack    []string
	}{
		{
			name:              "without a sitemap",
			options:           []service_config.Option{service_config.WithRobotsTxt(true)},
			expectedToContain: []string{"User-Agent: *", "Disallow: /"},
			expectedToLack:    []string{"Sitemap:", "Googlebot"},
		},
		{
			name: "with a sitemap",
			options: []service_config.Option{
				service_config.WithRobotsTxt(true),
				service_config.WithSitemap(true),
				service_config.WithEndpoints(staticContentEndpoint("/", "text/html")),
			},
			expectedToContain: []string{
				"User-Agent: *",
				"Disallow: /",
				"Googlebot",
				"Disallow: /api/",
				"Sitemap: https://example.com/sitemap.xml",
			},
		},
		{
			name: "with a sitemap of nothing",
			options: []service_config.Option{
				service_config.WithRobotsTxt(true),
				service_config.WithSitemap(true),
			},
			expectedToContain: []string{"User-Agent: *", "Disallow: /"},
			// Nothing is served that a crawler would index, so none is invited.
			expectedToLack: []string{"Sitemap:", "Googlebot"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			options := append([]service_config.Option{service_config.WithHost("example.com")}, testCase.options...)

			service, err := New(options...)
			if err != nil {
				t.Fatalf("new: %v", err)
			}

			robotsTxt := staticContentData(t, service.Mux, "/robots.txt")

			for _, expected := range testCase.expectedToContain {
				if !strings.Contains(robotsTxt, expected) {
					t.Errorf("robots.txt lacks %q:\n%s", expected, robotsTxt)
				}
			}

			for _, unexpected := range testCase.expectedToLack {
				if strings.Contains(robotsTxt, unexpected) {
					t.Errorf("robots.txt has %q:\n%s", unexpected, robotsTxt)
				}
			}
		})
	}
}

func TestSitemap(t *testing.T) {
	t.Parallel()

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithSitemap(true),
		service_config.WithEndpoints(
			staticContentEndpoint("/", "text/html"),
			staticContentEndpoint("/manual.pdf", "application/pdf"),
			staticContentEndpoint("/page.xhtml", "application/xhtml+xml"),
			staticContentEndpoint("/data.json", "application/json"),
		),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	sitemap := staticContentData(t, service.Mux, "/sitemap.xml")

	// Every document type a crawler indexes is listed, not only the HTML the mux calls a document.
	for _, expected := range []string{
		"<loc>https://example.com/</loc>",
		"<loc>https://example.com/manual.pdf</loc>",
		"<loc>https://example.com/page.xhtml</loc>",
	} {
		if !strings.Contains(sitemap, expected) {
			t.Errorf("the sitemap lacks %q:\n%s", expected, sitemap)
		}
	}

	if strings.Contains(sitemap, "data.json") {
		t.Errorf("the sitemap lists what no crawler indexes:\n%s", sitemap)
	}

	// The locations are sorted, so that the sitemap is the same from one start to the next despite
	// the endpoints being held in maps.
	if index := strings.Index(sitemap, "/manual.pdf"); index > strings.Index(sitemap, "/page.xhtml") {
		t.Errorf("the sitemap is not sorted:\n%s", sitemap)
	}
}

func TestSitemapWithoutHost(t *testing.T) {
	t.Parallel()

	// A sitemap wants absolute locations, which there is no deriving without a host.
	_, err := New(
		service_config.WithSitemap(true),
		service_config.WithEndpoints(staticContentEndpoint("/", "text/html")),
	)
	if err == nil {
		t.Error("new: got no error, want one")
	}
}

// TestSecurityTxt verifies that a service on a registered domain says how a vulnerability in it is
// reported without being told to, and without being told what to say.
func TestSecurityTxt(t *testing.T) {
	t.Parallel()

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithSecurityTxtContent(
			&motmedelHttpTypes.SecurityTxt{PreferredLanguages: []string{"sv", "en"}},
		),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	securityTxt := staticContentData(t, service.Mux, wellKnownSecurityTxtPath)

	for _, expected := range []string{
		"Contact: mailto:security@example.com",
		"Preferred-Languages: sv, en",
		// Filled in by the service: how long it is valid, and where it is expected to be found.
		"Expires: ",
		"Canonical: https://example.com/.well-known/security.txt",
	} {
		if !strings.Contains(securityTxt, expected) {
			t.Errorf("the security.txt lacks %q:\n%s", expected, securityTxt)
		}
	}

	// The path it had before RFC 9116 gave it a well-known one still leads there.
	redirect := service.Mux.Get(securityTxtPath, http.MethodGet)
	if redirect == nil {
		t.Fatalf("no endpoint at %q", securityTxtPath)
	}
	if redirect.Handler == nil {
		t.Errorf("the endpoint at %q does not redirect", securityTxtPath)
	}
}

func TestSecurityTxtExpiresIsKeptWhereConfigured(t *testing.T) {
	t.Parallel()

	expires := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithSecurityTxtContent(
			&motmedelHttpTypes.SecurityTxt{
				Contacts: []string{"mailto:security@elsewhere.example"},
				Expires:  expires,
			},
		),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	securityTxt := staticContentData(t, service.Mux, wellKnownSecurityTxtPath)

	// What is said outright is kept; only what is left out is derived.
	for _, expected := range []string{
		"Contact: mailto:security@elsewhere.example",
		"Expires: 2030-01-02T03:04:05Z",
	} {
		if !strings.Contains(securityTxt, expected) {
			t.Errorf("the security.txt lacks %q:\n%s", expected, securityTxt)
		}
	}
}

// TestSecurityTxtVariantFollowsTheHost verifies which of the two forms a service serves: its own
// where it is on a registered domain, a redirect to the registered domain's where it is on a
// subdomain, and none where there is no domain to derive either from.
func TestSecurityTxtVariantFollowsTheHost(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		host             string
		expectedContent  string
		expectedRedirect string
	}{
		{
			name:            "a registered domain",
			host:            "example.com",
			expectedContent: "Contact: mailto:security@example.com",
		},
		{
			name:            "a registered domain under a multi-part suffix",
			host:            "example.co.uk",
			expectedContent: "Contact: mailto:security@example.co.uk",
		},
		{
			name:             "a subdomain",
			host:             "service.example.com",
			expectedRedirect: "https://example.com/.well-known/security.txt",
		},
		{
			name:             "a deeper subdomain",
			host:             "a.b.example.com",
			expectedRedirect: "https://example.com/.well-known/security.txt",
		},
		// A development server has neither a domain to derive a contact from nor a reporter to
		// derive it for.
		{name: "localhost"},
		{name: "a subdomain of localhost", host: "service.localhost"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			host := testCase.host
			if host == "" {
				host = "localhost"
			}

			service, err := New(service_config.WithHost(host))
			if err != nil {
				t.Fatalf("new: %v", err)
			}

			endpoint := service.Mux.Get(wellKnownSecurityTxtPath, http.MethodGet)

			if testCase.expectedContent == "" && testCase.expectedRedirect == "" {
				if endpoint != nil {
					t.Error("a security.txt is served, though there is nothing to say in one")
				}
				return
			}

			if endpoint == nil {
				t.Fatal("no security.txt")
			}

			if expectedContent := testCase.expectedContent; expectedContent != "" {
				securityTxt := staticContentData(t, service.Mux, wellKnownSecurityTxtPath)
				if !strings.Contains(securityTxt, expectedContent) {
					t.Errorf("the security.txt lacks %q:\n%s", expectedContent, securityTxt)
				}
				return
			}

			if endpoint.StaticContent != nil {
				t.Error("the subdomain says a security.txt of its own")
			}
			if endpoint.Handler == nil {
				t.Fatal("the subdomain does not point at the registered domain's security.txt")
			}
		})
	}
}

func TestSecurityTxtTurnedOff(t *testing.T) {
	t.Parallel()

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithSecurityTxt(false),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	for _, path := range []string{securityTxtPath, wellKnownSecurityTxtPath} {
		if service.Mux.Get(path, http.MethodGet) != nil {
			t.Errorf("an endpoint is served at %q", path)
		}
	}
}

func TestSecurityTxtUrl(t *testing.T) {
	t.Parallel()

	securityTxtUrl := &url.URL{Scheme: "https", Host: "example.com", Path: wellKnownSecurityTxtPath}

	service, err := New(
		service_config.WithHost("service.example.com"),
		service_config.WithSecurityTxtUrl(securityTxtUrl),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	for _, path := range []string{securityTxtPath, wellKnownSecurityTxtPath} {
		endpoint := service.Mux.Get(path, http.MethodGet)
		if endpoint == nil {
			t.Fatalf("no endpoint at %q", path)
		}
		if endpoint.Handler == nil {
			t.Errorf("the endpoint at %q does not redirect", path)
		}
		if endpoint.StaticContent != nil {
			t.Errorf("the endpoint at %q serves a security.txt of its own", path)
		}
	}
}

func TestReporting(t *testing.T) {
	t.Parallel()

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithReporting(true),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	mux := service.Mux

	reportingEndpoints := mux.DefaultDocumentHeaders[reportingEndpointsHeaderName]
	for _, expected := range []string{cspReportToEndpointName, CspReportToPath, integrityEndpointName, IntegrityPath} {
		if !strings.Contains(reportingEndpoints, expected) {
			t.Errorf("%s lacks %q: %q", reportingEndpointsHeaderName, expected, reportingEndpoints)
		}
	}

	if integrityPolicy := mux.DefaultDocumentHeaders[integrityPolicyHeaderName]; !strings.Contains(integrityPolicy, integrityEndpointName) {
		t.Errorf("%s: got %q", integrityPolicyHeaderName, integrityPolicy)
	}

	csp, err := mux.GetContentSecurityPolicy()
	if err != nil {
		t.Fatalf("mux get content security policy: %v", err)
	}
	if csp == nil {
		t.Fatal("nil content security policy")
	}

	reportTo := csp.GetReportTo()
	if reportTo == nil || reportTo.Token != cspReportToEndpointName {
		t.Errorf("csp report-to: got %v, want %q", reportTo, cspReportToEndpointName)
	}

	// report-uri is deprecated and kept: it is the only way Firefox and Safari report at all.
	reportUri := csp.GetReportUri()
	if reportUri == nil || !strings.Contains(strings.Join(reportUri.UriReferences, " "), CspReportUriPath) {
		t.Errorf("csp report-uri: got %v, want %q", reportUri, CspReportUriPath)
	}

	for _, path := range []string{
		CspReportToPath,
		CspReportUriPath,
		IntegrityPath,
		JsErrorPath,
		JsUnhandledRejectionPath,
	} {
		if mux.Get(path, http.MethodPost) == nil {
			t.Errorf("no endpoint at %q", path)
		}
	}
}

// TestReportingDoesNotSupportNetworkErrorLogging verifies that nothing is answered with, nor served,
// for network error logging: it is reached only through the deprecated Report-To header, which is
// then not answered with either.
func TestReportingDoesNotSupportNetworkErrorLogging(t *testing.T) {
	t.Parallel()

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithProfile(service_config.ProfilePublicWeb),
		service_config.WithEndpoints(staticContentEndpoint("/", "text/html")),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	mux := service.Mux

	for _, headers := range []map[string]string{mux.DefaultHeaders, mux.DefaultDocumentHeaders} {
		for _, name := range []string{"NEL", "Report-To"} {
			if value, found := headers[name]; found {
				t.Errorf("%s: got %q, want no header", name, value)
			}
		}
	}

	if mux.Get("/api/report/network-error-logging", http.MethodPost) != nil {
		t.Error("a network error logging endpoint is served")
	}
}

// TestDoneCallbackLogsOncePerResponse verifies that a response served through a service that
// answers for a host is logged once. Both the mux and the vhost mux that fronts it call the
// callback, so a default on each would have the one response logged twice.
//
//nolint:paralleltest // The case sets the process-wide default logger, so it runs on its own.
func TestDoneCallbackLogsOncePerResponse(t *testing.T) {
	var buffer bytes.Buffer

	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelDebug})))

	service, err := New(
		service_config.WithHost("example.com"),
		service_config.WithEndpoints(noContentEndpoint()),
	)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	address := serveListener(t, service)
	doRequest(t, &http.Client{Transport: &http.Transport{}}, "http://"+address+"/", "example.com")

	if got := strings.Count(buffer.String(), "http_response_served"); got != 1 {
		t.Errorf("response served records: got %d, want 1\n%s", got, buffer.String())
	}
}
