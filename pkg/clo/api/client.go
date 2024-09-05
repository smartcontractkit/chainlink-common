// Package api implements an API client for the CLO GraphQL API.
//
// The api package is provided in the form of a library and intended to be imported
// by other packages or applications that want to communicate with the CLO API.
//
// The generated models and operations provided by this package are automatically
// generated and validated using a combination of [gqlgen] and [genqlient].
//
// [gqlgen]: https://github.com/99designs/gqlgen/
// [genqlient]: https://github.com/Khan/genqlient/
package api

import (
	"context"
	"net/url"
	"path"

	"github.com/Khan/genqlient/graphql"
	"github.com/go-playground/validator/v10"

	"github.com/smartcontractkit/feeds-manager/api/internal/auth"
	"github.com/smartcontractkit/feeds-manager/api/models"
	"github.com/smartcontractkit/feeds-manager/api/operations"
)

// Client is the public interface for the API client implementation.
type Client interface {
	// Ctx returns the Client's Context.
	Ctx() context.Context

	// Gql returns the Client's GraphQL client.
	Gql() graphql.Client

	// Login authenticates the Client's API session.
	Login() (*operations.LoginResponse, error)

	// Logout terminates the current API session.
	Logout() error
}

// client is the API Client for interacting with the backend provider.
type client struct {
	// context is the Context supplied to the API Client by the consumer.
	context context.Context

	// token is the session token used to authenticate with the API provider.
	token auth.Token

	// config is the provided configuration for the API Client.
	config *Config

	// graphQL is the GraphQL Client that handles the API requests/responses.
	graphQL graphql.Client
}

// Ctx returns the API Client's context
func (c *client) Ctx() context.Context {
	return c.context
}

// Gql returns the API Client's GraphQL client
func (c *client) Gql() graphql.Client {
	return c.graphQL
}

// Login authenticates the API Client using the credentials supplied in the
// config and stores the session token to use for subsequent API calls.
func (c *client) Login() (*operations.LoginResponse, error) {
	input := models.LoginInput{
		Email:    c.config.Email,
		Password: c.config.Password,
	}
	resp, err := operations.Login(c.context, c.graphQL, input)
	if err != nil {
		return nil, err
	}
	if resp.Login.Session != nil {
		if err = c.token.Save(resp.Login.Session.Token); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// Logout closes the current API session and clears the locally-cached token.
func (c *client) Logout() error {
	_, err := operations.Logout(c.context, c.graphQL)
	if err != nil {
		return err
	}
	err = c.token.Delete()
	if err != nil {
		return err
	}
	return nil
}

// Config describes the configuration of the API Client. It is intended to be read
// and parsed from a yaml file accessible to the application consuming this package.
type Config struct {
	// BaseURL is the base endpoint for the API.
	BaseUrl string `yaml:"BASE_URL" validate:"required,url"`

	// Email is the email address and/or username for the user's account.
	Email string `yaml:"EMAIL" validate:"required,email"`

	// Password is the password used to authenticate the user's account.
	Password string `yaml:"PASSWORD" validate:"required"`

	// FilePath is the path to the configuration file defining the Config.
	FilePath string `yaml:"-" validate:"required,file"`
}

// RequesEndpointUrl returns the fully formatted API request endpoint URL.
func (c *Config) requestEndpointUrl() string {
	// discard error because we validate the BaseUrl on Client initialization
	queryURL, _ := url.ParseRequestURI(c.BaseUrl)
	queryURL.Path = path.Join(queryURL.Path, "query")
	return queryURL.String()
}

// Validate returns an error if the Config is invalid or missing any fields.
func (c *Config) validate() error {
	v := validator.New()
	return v.Struct(c)
}

// NewClient initializes a new API client for the given context,
// using the values provided by the supplied Config. The Config is
// validated on Client initialization.
func NewClient(ctx context.Context, cfg *Config) (*client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	url := cfg.requestEndpointUrl()
	tkn := auth.NewTokenStore(cfg.FilePath)
	gql := auth.NewGqlClient(url, tkn)

	return &client{
		context: ctx,
		token:   tkn,
		config:  cfg,
		graphQL: gql,
	}, nil
}
