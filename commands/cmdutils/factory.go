// Forked from https://github.com/cli/cli/blob/929e082c13909044e2585af292ae952c9ca6f25c/pkg/cmd/factory/default.go
package cmdutils

import (
	"fmt"
	"net/http"
	"strings"

	graphql "github.com/hasura/go-graphql-client"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/internal/glrepo"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/glinstance"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
)

var (
	CachedConfig config.Config
	ConfigError  error
)

type Factory struct {
	HttpClient    func() (*gitlab.Client, error)
	BaseRepo      func() (glrepo.Interface, error)
	Remotes       func() (glrepo.Remotes, error)
	Config        func() (config.Config, error)
	Branch        func() (string, error)
	GraphQLClient func() (*graphql.Client, error)
	IO            *iostreams.IOStreams
}

func (f *Factory) RepoOverride(repo string) error {
	f.BaseRepo = func() (glrepo.Interface, error) {
		return glrepo.FromFullName(repo)
	}
	newRepo, err := f.BaseRepo()
	if err != nil {
		return err
	}
	// Initialise new http client for new repo host
	cfg, err := f.Config()
	if err == nil {
		OverrideAPIProtocol(cfg, newRepo)
	}
	f.HttpClient = func() (*gitlab.Client, error) {
		return LabClientFunc(newRepo.RepoHost(), cfg, false)
	}
	return nil
}

func LabClientFunc(repoHost string, cfg config.Config, isGraphQL bool) (*gitlab.Client, error) {
	c, err := api.NewClientWithCfg(repoHost, cfg, isGraphQL)
	if err != nil {
		return nil, err
	}
	return c.Lab(), nil
}

func remotesFunc() (glrepo.Remotes, error) {
	hostOverride := ""
	if !strings.EqualFold(glinstance.Default(), glinstance.OverridableDefault()) {
		hostOverride = glinstance.OverridableDefault()
	}
	rr := &remoteResolver{
		readRemotes: git.Remotes,
		getConfig:   configFunc,
	}
	fn := rr.Resolver(hostOverride)
	return fn()
}

func configFunc() (config.Config, error) {
	if CachedConfig != nil || ConfigError != nil {
		return CachedConfig, ConfigError
	}
	CachedConfig, ConfigError = initConfig()
	return CachedConfig, ConfigError
}

func baseRepoFunc() (glrepo.Interface, error) {
	remotes, err := remotesFunc()
	if err != nil {
		return nil, err
	}
	return remotes[0], nil
}

// OverrideAPIProtocol sets api protocol for host to initialize http client
func OverrideAPIProtocol(cfg config.Config, repo glrepo.Interface) {
	protocol, _ := cfg.Get(repo.RepoHost(), "api_protocol")
	api.SetProtocol(protocol)
}

func HTTPClientFactory(f *Factory) {
	f.HttpClient = func() (*gitlab.Client, error) {
		cfg, err := configFunc()
		if err != nil {
			return nil, err
		}
		repo, err := baseRepoFunc()
		if err != nil {
			// use default hostname if remote resolver fails
			repo = glrepo.NewWithHost("", "", glinstance.OverridableDefault())
		}
		OverrideAPIProtocol(cfg, repo)
		return LabClientFunc(repo.RepoHost(), cfg, false)
	}
}

func NewFactory() *Factory {
	return &Factory{
		Config:  configFunc,
		Remotes: remotesFunc,
		HttpClient: func() (*gitlab.Client, error) {
			// do not initialize httpclient since it may not be required by
			// some commands like version, help, etc...
			// It should be explicitly initialize with HTTPClientFactory()
			return nil, nil
		},
		BaseRepo: baseRepoFunc,
		Branch: func() (string, error) {
			currentBranch, err := git.CurrentBranch()
			if err != nil {
				return "", fmt.Errorf("could not determine current branch: %w", err)
			}
			return currentBranch, nil
		},
		IO: iostreams.Init(),
		GraphQLClient: func() (*graphql.Client, error) {

			cfg, err := configFunc()
			if err != nil {
				return nil, err
			}

			repo, err := baseRepoFunc()
			if err != nil {
				// use default hostname if remote resolver fails
				repo = glrepo.NewWithHost("", "", glinstance.OverridableDefault())
			}
			OverrideAPIProtocol(cfg, repo)

			c, err := api.NewClientWithCfg(repo.RepoHost(), cfg, true)
			if err != nil {
				return nil, err
			}

			client := c.HTTPClient()
			client.Transport = addTokenTransport{
				T:     client.Transport,
				Token: c.Token(),
			}
			url := glinstance.GraphQLEndpoint(repo.RepoHost(), c.Protocol)

			gqlClient := graphql.NewClient(url, c.HTTPClient())

			return gqlClient, nil
		},
	}
}

func initConfig() (config.Config, error) {
	if err := config.MigrateOldConfig(); err != nil {
		return nil, err
	}
	return config.Init()
}

type addTokenTransport struct {
	T     http.RoundTripper
	Token string
}

func (att addTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", att.Token))
	return att.T.RoundTrip(req)
}
