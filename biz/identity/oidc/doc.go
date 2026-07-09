// Package oidc bridges a standard OpenID Connect authorization-code flow into
// the post-auth identity mapping package.
//
// It handles provider metadata, authorization URLs, code exchange, ID-token
// verification, nonce checking, optional one-time login state storage, and
// assertion construction. It does not issue application sessions, cookies, or
// account-management screens.
package oidc
