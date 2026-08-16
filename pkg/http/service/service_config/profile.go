package service_config

// Profile is the kind of service being set up. It decides what a service answers with beyond what
// it was written to answer, along two axes: whether a browser renders documents from it, and
// whether what it serves is meant to be indexed.
//
// Internal here means behind authentication, reached by the users allowed in and by no one else. A
// crawler is not one of them, so an internal service says nothing to crawlers -- there is nothing
// they can reach to index, and nothing to keep them from.
//
// The users allowed in may be external people, though, and one of them may find a vulnerability. A
// security.txt is therefore worth serving whatever the profile, which is the other reason it is
// configured rather than decided by one.
//
// A profile is a starting point rather than a verdict. Options applied after it override what it
// decided, since options are applied in the order they are given.
//
// What needs saying that cannot be derived is not part of a profile: a security.txt names a contact
// nothing can guess, so it is configured with WithSecurityTxt whatever the profile.
type Profile string

const (
	// ProfileInternalApi is a service that answers requests from programs, behind authentication.
	// It permits nothing in its content security policy, having no document to permit anything to.
	ProfileInternalApi Profile = "internal_api"

	// ProfilePublicApi is a service that answers requests from programs, open to whoever asks. It
	// tells crawlers to keep out: one that finds it would otherwise index what is not a document.
	ProfilePublicApi Profile = "public_api"

	// ProfileInternalWeb is a service a browser renders documents from, behind authentication. It
	// reports what the browser blocks, so that a policy that breaks the pages is noticed, and says
	// nothing to crawlers, which do not reach what it serves.
	ProfileInternalWeb Profile = "internal_web"

	// ProfilePublicWeb is a service a browser renders documents from, open to whoever asks:
	// reporting what the browser blocks, and telling crawlers what there is to index.
	ProfilePublicWeb Profile = "public_web"
)

// profileSettings is what a profile decides.
type profileSettings struct {
	strictTransportSecurity  bool
	apiContentSecurityPolicy bool
	robotsTxt                bool
	sitemap                  bool
	reporting                bool
}

// profileToSettings is the whole of what the profiles decide. Every one of them is hardened in
// transport; what a browser is told, and what a crawler that can reach it is told, is what tells
// them apart.
var profileToSettings = map[Profile]*profileSettings{
	ProfileInternalApi: {
		strictTransportSecurity:  true,
		apiContentSecurityPolicy: true,
	},
	ProfilePublicApi: {
		strictTransportSecurity:  true,
		apiContentSecurityPolicy: true,
		robotsTxt:                true,
	},
	ProfileInternalWeb: {
		strictTransportSecurity: true,
		reporting:               true,
	},
	ProfilePublicWeb: {
		strictTransportSecurity: true,
		robotsTxt:               true,
		sitemap:                 true,
		reporting:               true,
	},
}

// Profiles returns the profiles that are defined, for a caller that lets its own configuration name
// one.
func Profiles() []Profile {
	return []Profile{ProfileInternalApi, ProfilePublicApi, ProfileInternalWeb, ProfilePublicWeb}
}

// IsValid reports whether the profile is one this package defines.
func (profile Profile) IsValid() bool {
	_, found := profileToSettings[profile]
	return found
}

func (profile Profile) String() string {
	return string(profile)
}

// WithProfile applies what the profile decides. An unknown profile is left to be reported by the
// service being made, rather than silently deciding nothing.
func WithProfile(profile Profile) Option {
	return func(config *Config) {
		config.Profile = profile

		settings, found := profileToSettings[profile]
		if !found || settings == nil {
			return
		}

		config.StrictTransportSecurity = settings.strictTransportSecurity
		config.ApiContentSecurityPolicy = settings.apiContentSecurityPolicy
		config.RobotsTxt = settings.robotsTxt
		config.Sitemap = settings.sitemap
		config.Reporting = settings.reporting
	}
}
