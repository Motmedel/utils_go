// Package targetpolicy centralizes network target safety checks for scanners
// and other components that initiate outbound connections to user-controlled
// or datastore-controlled targets.
//
// The policy rejects loopback, private, link-local, cloud metadata, reserved,
// documentation, multicast, and other non-routable address ranges before a
// target is scanned. ValidateTarget accepts raw IP addresses, hostnames,
// host:port values, and URLs; hostnames are resolved and every DNS answer must
// be outside the denied ranges. DNS lookup failures and empty DNS answers fail
// closed.
//
// Callers that use a scanner engine with its own connection policy can also
// pass DeniedPrefixStrings to that engine so connect-time DNS rebinding attempts
// are rejected by the same shared policy.
package targetpolicy
