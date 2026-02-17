package schema

import (
	"github.com/Motmedel/utils_go/pkg/net/types/domain_parts"
)

type HttpHeaders struct {
	Normalized string `json:"normalized,omitempty"`
	Original   string `json:"original,omitempty"`
}

type Body struct {
	Bytes   int    `json:"bytes,omitempty"`
	Content string `json:"content,omitempty"`
}

type Target struct {
	domain_parts.Parts
	Address string            `json:"address,omitempty"`
	Bytes   int               `json:"bytes,omitempty"`
	Domain  string            `json:"domain,omitempty"`
	Ip      string            `json:"ip,omitempty"`
	Mac     string            `json:"mac,omitempty"`
	Nat     *Nat              `json:"nat,omitempty"`
	Packets int               `json:"packets,omitempty"`
	Port    int               `json:"port,omitempty"`
	As      *AutonomousSystem `json:"as,omitempty"`
	Geo     *Geo              `json:"geo,omitempty"`
	User    *User             `json:"user,omitempty"`
}

type Base struct {
	Timestamp string            `json:"@timestamp,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Message   string            `json:"message,omitempty"`
	Tags      []string          `json:"tags,omitempty"`

	Client        *Target        `json:"client,omitempty"`
	Cloud         *Cloud         `json:"cloud,omitempty"`
	Destination   *Target        `json:"destination,omitempty"`
	Dns           *Dns           `json:"dns,omitempty"`
	Email         *Email         `json:"email,omitempty"`
	Error         *Error         `json:"error,omitempty"`
	Event         *Event         `json:"event,omitempty"`
	File          *File          `json:"file,omitempty"`
	Group         *Group         `json:"group,omitempty"`
	Host          *Host          `json:"host,omitempty"`
	Http          *Http          `json:"http,omitempty"`
	Log           *Log           `json:"log,omitempty"`
	Observer      *Observer      `json:"observer,omitempty"`
	Process       *Process       `json:"process,omitempty"`
	Registry      *Registry      `json:"registry,omitempty"`
	Related       *Related       `json:"related,omitempty"`
	Rule          *Rule          `json:"rule,omitempty"`
	Server        *Target        `json:"server,omitempty"`
	Source        *Target        `json:"source,omitempty"`
	Threat        *Threat        `json:"threat,omitempty"`
	Tls           *Tls           `json:"tls,omitempty"`
	Network       *Network       `json:"network,omitempty"`
	Url           *Url           `json:"url,omitempty"`
	User          *User          `json:"user,omitempty"`
	UserAgent     *UserAgent     `json:"user_agent,omitempty"`
	Vulnerability *Vulnerability `json:"vulnerability,omitempty"`

	// NOTE: Custom namespaces
	Whois *Whois `json:"whois,omitempty"`
	Tcp   *Tcp   `json:"tcp,omitempty"`
}

type AgentBuild struct {
	Original string `json:"original,omitempty"`
}

type Agent struct {
	Build       *AgentBuild `json:"build,omitempty"`
	EphemeralId string      `json:"ephemeral_id,omitempty"`
	Id          string      `json:"id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Type        string      `json:"type,omitempty"`
	Version     string      `json:"version,omitempty"`
}

type AutonomousSystem struct {
	Number       int64         `json:"number,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
}

type Geo struct {
	CityName       string `json:"city_name,omitempty"`
	ContinentCode  string `json:"continent_code,omitempty"`
	ContinentName  string `json:"continent_name,omitempty"`
	CountryIsoCode string `json:"country_iso_code,omitempty"`
	CountryName    string `json:"country_name,omitempty"`
	Location       any    `json:"location,omitempty"`
	Name           string `json:"name,omitempty"`
	PostalCode     string `json:"postal_code,omitempty"`
	RegionIsoCode  string `json:"region_iso_code,omitempty"`
	RegionName     string `json:"region_name,omitempty"`
	Timezone       string `json:"timezone,omitempty"`
}

type Nat struct {
	Ip   string `json:"ip,omitempty"`
	Port int    `json:"port,omitempty"`
}

type Group struct {
	Domain string `json:"domain,omitempty"`
	Id     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
}

type User struct {
	Domain   string   `json:"domain,omitempty"`
	Email    string   `json:"email,omitempty"`
	FullName string   `json:"full_name,omitempty"`
	Hash     string   `json:"hash,omitempty"`
	Id       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	Changes  *User    `json:"changes,omitempty"`
	Group    *Group   `json:"group,omitempty"`
	Target   *User    `json:"target,omitempty"`
}

type CloudAccount struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type CloudInstance struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type CloudMachine struct {
	Type string `json:"type,omitempty"`
}

type CloudProject struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type CloudService struct {
	Name string `json:"name,omitempty"`
}

type CloudOriginTargetAccount struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type CloudOriginTargetInstance struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type CloudOriginTargetMachine struct {
	Type string `json:"type,omitempty"`
}

type CloudOriginTargetProject struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type CloudOriginTargetService struct {
	Name string `json:"name,omitempty"`
}

type CloudOriginTarget struct {
	Account          *CloudOriginTargetAccount  `json:"account,omitempty"`
	AvailabilityZone string                     `json:"availability_zone,omitempty"`
	Instance         *CloudOriginTargetInstance `json:"instance,omitempty"`
	Machine          *CloudOriginTargetMachine  `json:"machine,omitempty"`
	Project          *CloudOriginTargetProject  `json:"project,omitempty"`
	Provider         string                     `json:"provider,omitempty"`
	Region           string                     `json:"region,omitempty"`
	Service          *CloudOriginTargetService  `json:"service,omitempty"`
}

type Cloud struct {
	Account          *CloudAccount      `json:"account,omitempty"`
	AvailabilityZone string             `json:"availability_zone,omitempty"`
	Instance         *CloudInstance     `json:"instance,omitempty"`
	Machine          *CloudMachine      `json:"machine,omitempty"`
	Origin           *CloudOriginTarget `json:"origin,omitempty"`
	Project          *CloudProject      `json:"project,omitempty"`
	Provider         string             `json:"provider,omitempty"`
	Region           string             `json:"region,omitempty"`
	Service          *CloudService      `json:"service,omitempty"`
	Target           *CloudOriginTarget `json:"target,omitempty"`
}

type Container struct {
	Id    string `json:"id,omitempty"`
	Image struct {
		Name string `json:"name,omitempty"`
		Tag  string `json:"tag,omitempty"`
	} `json:"image,omitempty"`
	Labels  any    `json:"labels,omitempty"`
	Name    string `json:"name,omitempty"`
	Runtime string `json:"runtime,omitempty"`
}

type DnsAnswer struct {
	Class string `json:"class,omitempty"`
	Data  string `json:"data,omitempty"`
	Name  string `json:"name,omitempty"`
	Ttl   int    `json:"ttl,omitempty"`
	Type  string `json:"type,omitempty"`
}

type DnsQuestion struct {
	domain_parts.Parts
	Class string `json:"class,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
}

type Dns struct {
	Answers      []*DnsAnswer `json:"answers,omitempty"`
	HeaderFlags  []string     `json:"header_flags,omitempty"`
	Id           string       `json:"id,omitempty"`
	OpCode       string       `json:"op_code,omitempty"`
	Question     *DnsQuestion `json:"question,omitempty"`
	ResolvedIp   []string     `json:"resolved_ip,omitempty"`
	ResponseCode string       `json:"response_code,omitempty"`
	Type         string       `json:"type,omitempty"`
}

type Error struct {
	// NOTE: Custom
	Cause      *Error `json:"cause,omitempty"`
	Code       string `json:"code,omitempty"`
	Id         string `json:"id,omitempty"`
	Message    string `json:"message,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
	Type       string `json:"type,omitempty"`
}

type EmailAttachmentFile struct {
	Extension string `json:"extension,omitempty"`
	MimeType  string `json:"mime_type,omitempty"`
	Name      string `json:"name,omitempty"`
	Size      int    `json:"size,omitempty"`
	Hash      *Hash  `json:"hash,omitempty"`
}

type EmailAttachment struct {
	File *EmailAttachmentFile `json:"file,omitempty"`
}

type EmailAddress struct {
	// NOTE: Custom
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
}

type Email struct {
	Attachments          []*EmailAttachment `json:"attachments,omitempty"`
	Bcc                  []*EmailAddress    `json:"bcc,omitempty"`
	Cc                   []*EmailAddress    `json:"cc,omitempty"`
	ContentType          string             `json:"content_type,omitempty"`
	DeliveryTimestamp    string             `json:"delivery_timestamp,omitempty"`
	Direction            string             `json:"direction,omitempty"`
	From                 []*EmailAddress    `json:"from,omitempty"`
	LocalId              string             `json:"local_id,omitempty"`
	MessageId            string             `json:"message_id,omitempty"`
	OriginationTimestamp string             `json:"origination_timestamp,omitempty"`
	ReplyTo              []*EmailAddress    `json:"reply_to,omitempty"`
	Sender               *EmailAddress      `json:"sender,omitempty"`
	Subject              string             `json:"subject,omitempty"`
	To                   []*EmailAddress    `json:"to,omitempty"`
	XMailer              string             `json:"x_mailer,omitempty"`
}

type Event struct {
	Action        string   `json:"action,omitempty"`
	AgentIdStatus string   `json:"agent_id_status,omitempty"`
	Category      []string `json:"category,omitempty"`
	Code          string   `json:"code,omitempty"`
	Created       string   `json:"created,omitempty"`
	Dataset       string   `json:"dataset,omitempty"`
	Duration      int64    `json:"duration,omitempty"`
	End           string   `json:"end,omitempty"`
	Hash          string   `json:"hash,omitempty"`
	Id            string   `json:"id,omitempty"`
	Ingested      string   `json:"ingested,omitempty"`
	Kind          string   `json:"kind,omitempty"`
	Module        string   `json:"module,omitempty"`
	Original      string   `json:"original,omitempty"`
	Outcome       string   `json:"outcome,omitempty"`
	Provider      string   `json:"provider,omitempty"`
	Reason        string   `json:"reason,omitempty"`
	Reference     string   `json:"reference,omitempty"`
	RiskScore     float64  `json:"risk_score,omitempty"`
	Sequence      int      `json:"sequence,omitempty"`
	Severity      int      `json:"severity,omitempty"`
	Start         string   `json:"start,omitempty"`
	Timezone      string   `json:"timezone,omitempty"`
	Type          []string `json:"type,omitempty"`
	Url           string   `json:"url,omitempty"`
}

type Hash struct {
	Cdhash string `json:"cdhash,omitempty"`
	Md5    string `json:"md5,omitempty"`
	Sha1   string `json:"sha1,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
	Sha384 string `json:"sha384,omitempty"`
	Sha512 string `json:"sha512,omitempty"`
	Ssdeep string `json:"ssdeep,omitempty"`
	Tlsh   string `json:"tlsh,omitempty"`
}

type File struct {
	Accessed    string   `json:"accessed,omitempty"`
	Attributes  []string `json:"attributes,omitempty"`
	Created     string   `json:"created,omitempty"`
	Ctime       string   `json:"ctime,omitempty"`
	Device      string   `json:"device,omitempty"`
	Directory   string   `json:"directory,omitempty"`
	DriveLetter string   `json:"drive_letter,omitempty"`
	Extension   string   `json:"extension,omitempty"`
	Gid         string   `json:"gid,omitempty"`
	Group       string   `json:"group,omitempty"`
	Hash        *Hash    `json:"hash,omitempty"`
	Inode       string   `json:"inode,omitempty"`
	Mode        string   `json:"mode,omitempty"`
	Mtime       string   `json:"mtime,omitempty"`
	Name        string   `json:"name,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	Path        string   `json:"path,omitempty"`
	Size        int64    `json:"size,omitempty"`
	Type        string   `json:"type,omitempty"`
	Uid         string   `json:"uid,omitempty"`
}

type Host struct {
	Architecture string   `json:"architecture,omitempty"`
	Domain       string   `json:"domain,omitempty"`
	Hostname     string   `json:"hostname,omitempty"`
	Id           string   `json:"id,omitempty"`
	Ip           []string `json:"ip,omitempty"`
	Mac          []string `json:"mac,omitempty"`
	Name         string   `json:"name,omitempty"`
	Type         string   `json:"type,omitempty"`
	Uptime       int64    `json:"uptime,omitempty"`
	Os           Os       `json:"os,omitempty"`
}

type HttpRequest struct {
	Body  *Body `json:"body,omitempty"`
	Bytes int   `json:"bytes,omitempty"`
	// NOTE: Custom
	ContentType string `json:"content_type,omitempty"`
	// NOTE: Custom
	HttpHeaders *HttpHeaders `json:"headers,omitempty"`
	Id          string       `json:"id,omitempty"`
	Method      string       `json:"method,omitempty"`
	MimeType    string       `json:"mime_type,omitempty"`
	Referrer    string       `json:"referrer,omitempty"`
}

type HttpResponse struct {
	Body  *Body `json:"body,omitempty"`
	Bytes int   `json:"bytes,omitempty"`
	// NOTE: Custom
	ContentType string `json:"content_type,omitempty"`
	// NOTE: Custom
	HttpHeaders *HttpHeaders `json:"headers,omitempty"`
	MimeType    string       `json:"mime_type,omitempty"`
	// NOTE: Custom
	ReasonPhrase string `json:"reason_phrase,omitempty"`
	StatusCode   int    `json:"status_code,omitempty"`
}

type Http struct {
	Request  *HttpRequest  `json:"request,omitempty"`
	Response *HttpResponse `json:"response,omitempty"`
	Version  string        `json:"version,omitempty"`
}

type Interface struct {
	Alias string `json:"alias,omitempty"`
	Id    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
}

type Network struct {
	Application string `json:"application,omitempty"`
	Bytes       int64  `json:"bytes,omitempty"`
	// NOTE: Made into an array to be able to capture both Source-Destination and Client-Server, which can differ.
	CommunityId []string `json:"community_id,omitempty"`
	Direction   string   `json:"direction,omitempty"`
	ForwardedIp string   `json:"forwarded_ip,omitempty"`
	IanaNumber  string   `json:"iana_number,omitempty"`
	Inner       any      `json:"inner,omitempty"`
	Name        string   `json:"name,omitempty"`
	Packets     int64    `json:"packets,omitempty"`
	Protocol    string   `json:"protocol,omitempty"`
	Transport   string   `json:"transport,omitempty"`
	Type        string   `json:"type,omitempty"`
}

type LogOriginFile struct {
	Line int    `json:"line,omitempty"`
	Name string `json:"name,omitempty"`
}

type LogFile struct {
	Path string `json:"path,omitempty"`
}

type LogOrigin struct {
	File     *LogOriginFile `json:"file,omitempty"`
	Function string         `json:"function,omitempty"`
	// NOTE: Custom
	Process *Process `json:"process,omitempty"`
}

type LogSyslogFacility struct {
	Code int    `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type LogSyslogSeverity struct {
	Code int    `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type LogSyslog struct {
	Appname        string             `json:"appname,omitempty"`
	Facility       *LogSyslogFacility `json:"facility,omitempty"`
	Hostname       string             `json:"hostname,omitempty"`
	Msgid          string             `json:"msgid,omitempty"`
	Priority       int                `json:"priority,omitempty"`
	Procid         string             `json:"procid,omitempty"`
	StructuredData map[string]any     `json:"structured_data,omitempty"`
	Version        string             `json:"version,omitempty"`
}

type Log struct {
	LogFile *LogFile   `json:"file,omitempty"`
	Level   string     `json:"level,omitempty"`
	Logger  string     `json:"logger,omitempty"`
	Origin  *LogOrigin `json:"origin,omitempty"`
	Syslog  *LogSyslog `json:"syslog,omitempty"`
}

type ObserverIngressEgress struct {
	Interface *Interface `json:"interface,omitempty"`
	Zone      string     `json:"zone,omitempty"`
}

type Observer struct {
	Egress *ObserverIngressEgress `json:"egress,omitempty"`
	// NOTE: Custom
	Hook         string                 `json:"hook,omitempty"`
	Hostname     string                 `json:"hostname,omitempty"`
	Ingress      *ObserverIngressEgress `json:"ingress,omitempty"`
	Ip           string                 `json:"ip,omitempty"`
	Mac          string                 `json:"mac,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Os           *Os                    `json:"os,omitempty"`
	Product      string                 `json:"product,omitempty"`
	SerialNumber string                 `json:"serial_number,omitempty"`
	Type         string                 `json:"type,omitempty"`
	Vendor       string                 `json:"vendor,omitempty"`
	Version      string                 `json:"version,omitempty"`
}

type Organization struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Os struct {
	Family   string `json:"family,omitempty"`
	Full     string `json:"full,omitempty"`
	Kernel   string `json:"kernel,omitempty"`
	Name     string `json:"name,omitempty"`
	Platform string `json:"platform,omitempty"`
	Type     string `json:"type,omitempty"`
	Version  string `json:"version,omitempty"`
}

type ProcessIo struct {
	Text string `json:"text,omitempty"`
}

type ProcessThreadCapabilities struct {
	Effective []string `json:"effective,omitempty"`
	Permitted []string `json:"permitted,omitempty"`
}

type ProcessThread struct {
	Id   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Process struct {
	Args             []string       `json:"args,omitempty"`
	ArgsCount        int            `json:"args_count,omitempty"`
	CommandLine      string         `json:"command_line,omitempty"`
	End              string         `json:"end,omitempty"`
	EnvVars          []string       `json:"env_vars,omitempty"`
	Executable       string         `json:"executable,omitempty"`
	ExitCode         int            `json:"exit_code,omitempty"`
	Group            *Group         `json:"group,omitempty"`
	Interactive      bool           `json:"interactive,omitempty"`
	Io               *ProcessIo     `json:"io,omitempty"`
	Name             string         `json:"name,omitempty"`
	Parent           *Process       `json:"parent,omitempty"`
	Pid              int            `json:"pid,omitempty"`
	Previous         *Process       `json:"previous,omitempty"`
	Start            string         `json:"start,omitempty"`
	Thread           *ProcessThread `json:"thread,omitempty"`
	Title            string         `json:"title,omitempty"`
	Uptime           int            `json:"uptime,omitempty"`
	User             *User          `json:"user,omitempty"`
	WorkingDirectory string         `json:"working_directory,omitempty"`
}

type RegistryData struct {
	Bytes   string   `json:"bytes,omitempty"`
	Strings []string `json:"strings,omitempty"`
	Type    string   `json:"type,omitempty"`
}

type Registry struct {
	Data  *RegistryData `json:"data,omitempty"`
	Hive  string        `json:"hive,omitempty"`
	Key   string        `json:"key,omitempty"`
	Path  string        `json:"path,omitempty"`
	Value string        `json:"value,omitempty"`
}

type Related struct {
	Hash  []string `json:"hash,omitempty"`
	Hosts []string `json:"hosts,omitempty"`
	Ip    []string `json:"ip,omitempty"`
	User  []string `json:"user,omitempty"`
}

type Rule struct {
	Author      string `json:"author,omitempty"`
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
	Id          string `json:"id,omitempty"`
	License     string `json:"license,omitempty"`
	Name        string `json:"name,omitempty"`
	Reference   string `json:"reference,omitempty"`
	Ruleset     string `json:"ruleset,omitempty"`
	UUID        string `json:"uuid,omitempty"`
	Version     string `json:"version,omitempty"`
}

type Service struct {
	Id    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	State string `json:"state,omitempty"`
	Type  string `json:"type,omitempty"`
}

// NOTE: Custom

type Tcp struct {
	Flags                 []string `json:"flags,omitempty"`
	AcknowledgementNumber *int     `json:"acknowledgement_number,omitempty"`
	SequenceNumber        *int     `json:"sequence_number,omitempty"`
	State                 string   `json:"state,omitempty"`
}

type ThreatGroup struct {
	Alias     []string `json:"alias,omitempty"`
	Id        string   `json:"id,omitempty"`
	Name      string   `json:"name,omitempty"`
	Reference string   `json:"reference,omitempty"`
}

type ThreatFeed struct {
	Name      string `json:"name,omitempty"`
	Reference string `json:"reference,omitempty"`
}

type ThreatTechnique struct {
	Id           []string         `json:"id,omitempty"`
	Name         []string         `json:"name,omitempty"`
	Reference    []string         `json:"reference,omitempty"`
	Subtechnique *ThreatTechnique `json:"subtechnique,omitempty"`
}

type ThreatTactic struct {
	Id        []string `json:"id,omitempty"`
	Name      []string `json:"name,omitempty"`
	Reference []string `json:"reference,omitempty"`
}

type ThreatSoftware struct {
	Alias     string   `json:"alias,omitempty"`
	Id        string   `json:"id,omitempty"`
	Name      string   `json:"name,omitempty"`
	Platforms []string `json:"platforms,omitempty"`
	Reference string   `json:"reference,omitempty"`
	Type      string   `json:"type,omitempty"`
}

type ThreatIndicatorMarking struct {
	Tlp        string `json:"tlp,omitempty"`
	TlpVersion string `json:"tlp_version,omitempty"`
}

type ThreatIndicatorEmail struct {
	Address string `json:"address,omitempty"`
}

type ThreatIndicator struct {
	Confidence   string                  `json:"confidence,omitempty"`
	Description  string                  `json:"description,omitempty"`
	Email        *ThreatIndicatorEmail   `json:"email,omitempty"`
	File         *File                   `json:"file,omitempty"`
	Geo          *Geo                    `json:"geo,omitempty"`
	FirstSeen    string                  `json:"first_seen,omitempty"`
	Id           string                  `json:"id,omitempty"`
	Ip           string                  `json:"ip,omitempty"`
	LastSeen     string                  `json:"last_seen,omitempty"`
	Marking      *ThreatIndicatorMarking `json:"marking,omitempty"`
	ModifiedAt   string                  `json:"modified_at,omitempty"`
	Name         string                  `json:"name,omitempty"`
	Port         *int                    `json:"port,omitempty"`
	Provider     string                  `json:"provider,omitempty"`
	Reference    string                  `json:"reference,omitempty"`
	Registry     *Registry               `json:"registry,omitempty"`
	ScannerStats *int                    `json:"scanner_stats,omitempty"`
	Sightings    *int                    `json:"sightings,omitempty"`
	Type         string                  `json:"type,omitempty"`
	Url          *Url                    `json:"url,omitempty"`
}

type ThreatEnrichmentMatched struct {
	Atomic   string `json:"atomic,omitempty"`
	Field    string `json:"field,omitempty"`
	Id       string `json:"id,omitempty"`
	Index    string `json:"index,omitempty"`
	Occurred string `json:"occurred,omitempty"`
	Type     string `json:"type,omitempty"`
}

type ThreatEnrichment struct {
	Indicator *ThreatIndicator         `json:"indicator,omitempty"`
	Matched   *ThreatEnrichmentMatched `json:"matched,omitempty"`
}

type Threat struct {
	Enrichments []*ThreatEnrichment `json:"enrichments,omitempty"`
	Feed        *ThreatFeed         `json:"feed,omitempty"`
	Framework   string              `json:"framework,omitempty"`
	Group       *ThreatGroup        `json:"group,omitempty"`
	Indicator   *ThreatIndicator    `json:"indicator,omitempty"`
	Software    *ThreatSoftware     `json:"software,omitempty"`
	Tactic      *ThreatTactic       `json:"tactic,omitempty"`
	Technique   *ThreatTechnique    `json:"technique,omitempty"`
}

type TlsHash struct {
	Md5    string `json:"md5,omitempty"`
	Sha1   string `json:"sha1,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

type TlsClient struct {
	Certificate      string   `json:"certificate,omitempty"`
	CertificateChain []string `json:"certificate_chain,omitempty"`
	Hash             *TlsHash `json:"hash,omitempty"`
	Issuer           string   `json:"issuer,omitempty"`
	Ja3              string   `json:"ja3,omitempty"`
	NotAfter         string   `json:"not_after,omitempty"`
	NotBefore        string   `json:"not_before,omitempty"`
	ServerName       string   `json:"server_name,omitempty"`
	Subject          string   `json:"subject,omitempty"`
	SupportedCiphers []string `json:"supported_ciphers,omitempty"`
	X509             *X509    `json:"x509,omitempty"`
}

type TlsServer struct {
	Certificate      string   `json:"certificate,omitempty"`
	CertificateChain []string `json:"certificate_chain,omitempty"`
	Hash             *TlsHash `json:"hash,omitempty"`
	Issuer           string   `json:"issuer,omitempty"`
	Ja3s             string   `json:"ja3s,omitempty"`
	NotAfter         string   `json:"not_after,omitempty"`
	NotBefore        string   `json:"not_before,omitempty"`
	Subject          string   `json:"subject,omitempty"`
	X509             *X509    `json:"x509,omitempty"`
}

// NOTE: Custom/OpenTelemetry

type TlsProtocol struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type Tls struct {
	Cipher          string       `json:"cipher,omitempty"`
	Client          *TlsClient   `json:"client,omitempty"`
	Curve           string       `json:"curve,omitempty"`
	Established     bool         `json:"established,omitempty"`
	NextProtocol    string       `json:"next_protocol,omitempty"`
	Resumed         bool         `json:"resumed,omitempty"`
	Server          *TlsServer   `json:"server,omitempty"`
	TlsProtocol     *TlsProtocol `json:"protocol,omitempty"`
	Version         string       `json:"version,omitempty"`
	VersionProtocol string       `json:"version_protocol,omitempty"`
}

type Url struct {
	domain_parts.Parts
	Domain    string `json:"domain,omitempty"`
	Extension string `json:"extension,omitempty"`
	Fragment  string `json:"fragment,omitempty"`
	Full      string `json:"full,omitempty"`
	Original  string `json:"original,omitempty"`
	Password  string `json:"password,omitempty"`
	Path      string `json:"path,omitempty"`
	Port      int    `json:"port,omitempty"`
	Query     string `json:"query,omitempty"`
	Scheme    string `json:"scheme,omitempty"`
	Username  string `json:"username,omitempty"`
}

type UserAgentDevice struct {
	Name string `json:"name,omitempty"`
}

type UserAgent struct {
	Device   *UserAgentDevice `json:"device,omitempty"`
	Name     string           `json:"name,omitempty"`
	Original string           `json:"original,omitempty"`
	Os       *Os              `json:"os,omitempty"`
	Version  string           `json:"version,omitempty"`
}

type VulnerabilityScanner struct {
	Vendor string `json:"vendor,omitempty"`
}

type VulnerabilityScore struct {
	Base          float64 `json:"base,omitempty"`
	Environmental float64 `json:"environmental,omitempty"`
	Temporal      float64 `json:"temporal,omitempty"`
	Version       string  `json:"version,omitempty"`
}

type Vulnerability struct {
	Category       string                `json:"category,omitempty"`
	Classification string                `json:"classification,omitempty"`
	Description    string                `json:"description,omitempty"`
	Enumeration    string                `json:"enumeration,omitempty"`
	Id             string                `json:"id,omitempty"`
	Reference      string                `json:"reference,omitempty"`
	ReportId       string                `json:"report_id,omitempty"`
	Scanner        *VulnerabilityScanner `json:"scanner,omitempty"`
	Score          *VulnerabilityScore   `json:"score,omitempty"`
	Severity       string                `json:"severity,omitempty"`
}

type X509Target struct {
	CommonName         []string `json:"common_name,omitempty"`
	Country            []string `json:"country,omitempty"`
	DistinguishedName  string   `json:"distinguished_name,omitempty"`
	Locality           []string `json:"locality,omitempty"`
	Organization       []string `json:"organization,omitempty"`
	OrganizationalUnit []string `json:"organizational_unit,omitempty"`
	StateOrProvince    []string `json:"state_or_province,omitempty"`
}

type X509 struct {
	AlternativeNames   []string    `json:"alternate_names,omitempty"`
	Issuer             *X509Target `json:"issuer,omitempty"`
	NotAfter           string      `json:"not_after,omitempty"`
	NotBefore          string      `json:"not_before,omitempty"`
	PublicKeyAlgorithm string      `json:"public_key_algorithm,omitempty"`
	PublicKeyCurve     string      `json:"public_key_curve,omitempty"`
	PublicKeyExponent  int         `json:"public_key_exponent,omitempty"`
	PublicKeySize      int         `json:"public_key_size,omitempty"`
	SerialNumber       string      `json:"serial_number,omitempty"`
	SignatureAlgorithm string      `json:"signature_algorithm,omitempty"`
	Subject            *X509Target `json:"subject,omitempty"`
	VersionNumber      string      `json:"version_number,omitempty"`
}

// NOTE: Custom

type WhoisRequest struct {
	Body *Body  `json:"body,omitempty"`
	Id   string `json:"id,omitempty"`
}

// NOTE: Custom

type WhoisResponse struct {
	Body *Body `json:"body,omitempty"`
}

// NOTE: Custom

type Whois struct {
	Request  *WhoisRequest  `json:"request,omitempty"`
	Response *WhoisResponse `json:"response,omitempty"`
}
