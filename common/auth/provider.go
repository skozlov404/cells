package auth

import (
	"net/url"
	"time"

	"github.com/ory/hydra/driver/configuration"
	"github.com/ory/hydra/x"
	"github.com/ory/x/tracing"
	"github.com/rs/cors"

	"github.com/pydio/cells/common"
	"github.com/pydio/cells/common/config"
	"github.com/pydio/cells/common/utils/std"
)

type ConfigurationProvider interface {
	configuration.Provider
	Clients() common.Scanner
	Connectors() common.Scanner
}

type configurationProvider struct {
	// rootURL
	r string

	// values
	v common.ConfigValues

	cors       common.ConfigValues
	urls       common.ConfigValues
	oidc       common.ConfigValues
	clients    common.Scanner
	connectors common.Scanner

	drv string
	dsn string
}

var (
	conf                 ConfigurationProvider
	onConfigurationInits []func()
	confInit             bool
)

func InitConfiguration(values common.ConfigValues) {
	externalURL := config.Values("defaults/url").String()

	conf = NewProvider(externalURL, values)

	for _, onConfigurationInit := range onConfigurationInits {
		onConfigurationInit()
	}

	confInit = true
}

func OnConfigurationInit(f func()) {
	onConfigurationInits = append(onConfigurationInits, f)

	if confInit == true {
		f()
	}
}

func GetConfigurationProvider() ConfigurationProvider {
	return conf
}

func NewProvider(rootURL string, values common.ConfigValues) ConfigurationProvider {
	return &configurationProvider{
		r:          rootURL,
		v:          values,
		cors:       values.Values("cors"),
		urls:       values.Values("urls"),
		oidc:       values.Values("oidc"),
		clients:    values.Values("staticClients"),
		connectors: values.Values("connectors"),
	}
}

func (v *configurationProvider) InsecureRedirects() []string {
	return v.v.Values("insecureRedirects").StringArray()
}

func (v *configurationProvider) WellKnownKeys(include ...string) []string {
	if v.AccessTokenStrategy() == "jwt" {
		include = append(include, x.OAuth2JWTKeyName)
	}

	include = append(include, x.OpenIDConnectKeyName)

	return include
}

func (v *configurationProvider) ServesHTTPS() bool {
	return v.v.Values("https").Bool()
}

func (v *configurationProvider) IsUsingJWTAsAccessTokens() bool {
	return v.AccessTokenStrategy() != "opaque"
}

func (v *configurationProvider) SubjectTypesSupported() []string {
	return v.v.Values("subjectTypesSupported").Default([]string{"public"}).StringArray()
}

func (v *configurationProvider) DefaultClientScope() []string {
	return v.v.Values("defaultClientScope").Default([]string{"offline_access", "offline", "openid", "pydio", "email"}).StringArray()
}

func (v *configurationProvider) CORSEnabled(iface string) bool {
	return v.cors.Values(iface) != nil
}

func (v *configurationProvider) CORSOptions(iface string) cors.Options {
	return cors.Options{
		AllowedOrigins:     v.cors.Values(iface, "allowedOrigins").StringArray(),
		AllowedMethods:     v.cors.Values(iface, "allowedMethods").StringArray(),
		AllowedHeaders:     v.cors.Values(iface, "allowedHeaders").StringArray(),
		ExposedHeaders:     v.cors.Values(iface, "exposedHeaders").StringArray(),
		AllowCredentials:   v.cors.Values(iface, "allowCredentials").Default(true).Bool(),
		OptionsPassthrough: v.cors.Values(iface, "optionsPassthrough").Bool(),
		MaxAge:             v.cors.Values(iface, "maxAge").Int(),
		Debug:              v.cors.Values(iface, "debug").Bool(),
	}
}

func (v *configurationProvider) DSN() string {
	d := v.v.Values("dsn").Default(std.Reference("#/defaults/database")).StringMap()
	return d["drv"] + "://" + d["dsn"]
}

func (v *configurationProvider) DataSourcePlugin() string {
	d := v.v.Values("dsn").Default(std.Reference("#/defaults/database")).StringMap()
	return d["drv"] + "://" + d["dsn"]
}

func (v *configurationProvider) BCryptCost() int {
	return 10
}

func (v *configurationProvider) AdminListenOn() string {
	return ":0"
}

func (v *configurationProvider) AdminDisableHealthAccessLog() bool {
	return false
}

func (v *configurationProvider) PublicListenOn() string {
	return ":0"
}

func (v *configurationProvider) PublicDisableHealthAccessLog() bool {
	return v.v.Values("publicDisabledHealthAccessLog").Bool()
}

func (v *configurationProvider) ConsentRequestMaxAge() time.Duration {
	return v.v.Values("consentRequestMaxAge").Default(30 * time.Minute).Duration()
}

func (v *configurationProvider) AccessTokenLifespan() time.Duration {
	return v.v.Values("accessTokenLifespan").Default(10 * time.Minute).Duration()
}

func (v *configurationProvider) RefreshTokenLifespan() time.Duration {
	return v.v.Values("refreshTokenLifespan").Default(1 * time.Hour).Duration()
}

func (v *configurationProvider) IDTokenLifespan() time.Duration {
	return v.v.Values("idTokenLifespan").Default(1 * time.Hour).Duration()
}

func (v *configurationProvider) AuthCodeLifespan() time.Duration {
	return v.v.Values("authCodeLifespan").Default(10 * time.Minute).Duration()
}

func (v *configurationProvider) ScopeStrategy() string {
	return ""
}

func (v *configurationProvider) TracingServiceName() string {
	return "ORY Hydra"
}

func (v *configurationProvider) TracingProvider() string {
	return ""
}

func (v *configurationProvider) TracingJaegerConfig() *tracing.JaegerConfig {
	return &tracing.JaegerConfig{}
}

func (v *configurationProvider) GetCookieSecrets() [][]byte {
	return [][]byte{
		v.GetSystemSecret(),
	}
}

func (v *configurationProvider) GetRotatedSystemSecrets() [][]byte {
	secrets := [][]byte{v.GetSystemSecret()}

	if len(secrets) < 2 {
		return nil
	}

	var rotated [][]byte
	for _, secret := range secrets[1:] {
		rotated = append(rotated, x.HashByteSecret(secret))
	}

	return rotated
}

func (v *configurationProvider) GetSystemSecret() []byte {
	return []byte(v.v.Values("secret").String())
}

func (v *configurationProvider) LogoutRedirectURL() *url.URL {
	u, _ := url.Parse(v.r + v.urls.Values("logoutRedirectURL").Default("/oauth2/logout/callback").String())
	return u
}

func (v *configurationProvider) LoginURL() *url.URL {
	u, _ := url.Parse(v.r + v.urls.Values("loginURL").Default("/oauth2/login").String())
	return u
}

func (v *configurationProvider) LogoutURL() *url.URL {
	u, _ := url.Parse(v.r + v.urls.Values("logoutURL").Default("/oauth2/logout").String())
	return u
}

func (v *configurationProvider) ConsentURL() *url.URL {
	u, _ := url.Parse(v.r + v.urls.Values("consentURL").Default("/oauth2/consent").String())
	return u
}

func (v *configurationProvider) ErrorURL() *url.URL {
	u, _ := url.Parse(v.r + v.urls.Values("errorURL").Default("/oauth2/fallbacks/error").String())
	return u
}

func (v *configurationProvider) PublicURL() *url.URL {
	u, _ := url.Parse(v.r + v.urls.Values("publicURL").Default("/oidc/").String())
	return u
}

func (v *configurationProvider) IssuerURL() *url.URL {
	u, _ := url.Parse(v.r + v.urls.Values("issuerURL").Default("/oidc/").String())
	return u
}

func (v *configurationProvider) OAuth2AuthURL() string {
	return v.urls.Values("oauth2AuthURL").Default("/oauth2/auth").String() // this should not have the host etc prepended...
}

func (v *configurationProvider) OAuth2ClientRegistrationURL() *url.URL {
	u, _ := url.Parse(v.r + v.urls.Values("loginURL").Default("").String())
	return u
}

func (v *configurationProvider) AllowTLSTerminationFrom() []string {
	return v.v.Values("allowTLSTerminationFrom").Default([]string{}).StringArray()
}

func (v *configurationProvider) AccessTokenStrategy() string {
	return v.v.Values("accessTokenStrategy").Default("opaque").String()
}

func (v *configurationProvider) SubjectIdentifierAlgorithmSalt() string {
	return v.v.Values("subjectIdentifierAlgorithmSalt").Default("").String()
}

func (v *configurationProvider) OIDCDiscoverySupportedClaims() []string {
	return v.oidc.Values("supportedClaims").Default([]string{}).StringArray()
}

func (v *configurationProvider) OIDCDiscoverySupportedScope() []string {
	return v.oidc.Values("supportedScope").Default([]string{}).StringArray()
}

func (v *configurationProvider) OIDCDiscoveryUserinfoEndpoint() string {
	return v.oidc.Values("userInfoEndpoint").Default("/oauth2/userinfo").String()
}

func (v *configurationProvider) ShareOAuth2Debug() bool {
	return v.v.Values("shareOAuth2Debug").Default(false).Bool()
}

func (v *configurationProvider) Clients() common.Scanner {
	return v.clients
}

func (v *configurationProvider) Connectors() common.Scanner {
	return v.connectors
}
