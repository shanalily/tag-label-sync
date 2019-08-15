package azure

import (
	"os"

	"github.com/Azure/go-autorest/autorest/azure"
)

type AuthContext interface {
	ClientID() string
	ClientSecret() string
	TenantID() string

	Environment() string
}

type authContext struct {
	clientID        string
	clientSecret    string `json:clientId"`
	tenantID        string `json:clientSecret"`
	subscriptionID  string `json:tenantId"`
	locationDefault string
	cloudName       string
	baseGroupName   string
	userAgent       string
	environment     *azure.Environment
}

// func NewAuthContext() AuthContext {
//	ctx, _ :=
// }

func (ac *authContext) ClientID() string {
	return clientID
}

func (ac *authContext) ClientSecret() string {
	return clientSecret
}

func (ac *authContext) TenantID() string {
	return tenantID
}

func (ac *authContext) SubscriptionID() string {
	return subscriptionID
}

func (ac *authContext) DefaultLocation() string {
	return locationDefault
}

func (ac *authContext) UserAgent() string {
	return userAgent
}

func (ac *authContext) Environment() *azure.Environment {
	if environment != nil {
		return environment
	}
	env, err := azure.EnvironmentFromName(cloudName)
	if err != nil {
		os.Exit(1)
	}
	environment = &env
	return environment
}
