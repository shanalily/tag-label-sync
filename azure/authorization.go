package azure

import (
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
)

var (
	graphAuthorizer autorest.Authorizer
)

type client struct {
	authCtx authContext
}

// taken from azure-sdk-for-go samples
// GetGraphAuthorizer gets an OAuthTokenAuthorizer for graphrbac API.
func (c *client) GetGraphAuthorizer() (autorest.Authorizer, error) {
	if graphAuthorizer != nil {
		return graphAuthorizer, nil
	}

	var a autorest.Authorizer
	var err error

	a, err = c.getAuthorizerForResource(c.Environment().GraphEndpoint)

	if err == nil {
		// cache
		graphAuthorizer = a
	} else {
		graphAuthorizer = nil
	}

	return graphAuthorizer, err
}

func (c *client) getAuthorizerForResource(resource string) (autorest.Authorizer, error) {

	var a autorest.Authorizer
	var err error
	oauthConfig, err := adal.NewOAuthConfig(
		Environment().ActiveDirectoryEndpoint, TenantID())
	if err != nil {
		return nil, err
	}

	token, err := adal.NewServicePrincipalToken(
		*oauthConfig, ClientID(), ClientSecret(), resource)
	if err != nil {
		return nil, err
	}
	a = autorest.NewBearerAuthorizer(token)

	return a, err
}

func newServicePrincipalClient(tenantID string) graphrbac.ServicePrincipalsClient {
	client := graphrbac.NewServicePrincipalsClient(TenantID())
	// get graph authorizer
	// client.Authorizer = a
	client.AddToUserAgent(UserAgent())
	return client
}
