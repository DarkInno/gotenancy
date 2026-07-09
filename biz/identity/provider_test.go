package identity

import "testing"

func TestProviderPresetsValidate(t *testing.T) {
	tests := []Provider{
		GoogleOIDC(),
		GitHubOAuth(),
		MicrosoftEntraID("organizations"),
		MagicLink("stytch"),
		SAML("okta_saml", "https://idp.example.com/entity", "https://idp.example.com/sso"),
		GenericOIDC("auth0", "https://example.us.auth0.com/"),
	}

	for _, provider := range tests {
		if err := provider.Validate(); err != nil {
			t.Fatalf("Provider(%s).Validate() error = %v", provider.Key, err)
		}
	}
}

func TestMicrosoftEntraIDDefaultsToCommonTenant(t *testing.T) {
	provider := MicrosoftEntraID(" ")
	if provider.Issuer != "https://login.microsoftonline.com/common/v2.0" {
		t.Fatalf("MicrosoftEntraID issuer = %q, want common tenant issuer", provider.Issuer)
	}
}

func TestProviderCloneProtectsSlices(t *testing.T) {
	provider := GoogleOIDC()
	cloned := cloneProvider(provider)
	cloned.Scopes[0] = "mutated"

	if provider.Scopes[0] != "openid" {
		t.Fatalf("provider scopes mutated = %#v", provider.Scopes)
	}
}
