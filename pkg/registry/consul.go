package registry

import (
	confpkg "flamingo/pkg/config"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/google/wire"
	consulAPI "github.com/hashicorp/consul/api"
)

var ProviderSet = wire.NewSet(
	NewConsulRegistry,
)

func NewConsulRegistry(conf confpkg.Registry) *consul.Registry {
	c := consulAPI.DefaultConfig()
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	cli, err := consulAPI.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(false))
	return r
}
