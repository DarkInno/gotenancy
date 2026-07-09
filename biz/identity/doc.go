// Package identity maps verified external auth and SSO identities to
// GoTenancy users and tenant memberships.
//
// The package intentionally does not implement password authentication,
// OAuth callback handlers, ID-token validation, magic-link delivery, or SAML
// XML validation. Applications should verify provider assertions with their
// IdP SDK or protocol library first, then pass the verified identity into this
// package to create a tenant-scoped user/member link.
package identity
