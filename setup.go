package filterresponse

import (
    "net"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("filterresponse", setup) }

func setup(c *caddy.Controller) error {
	c.Next()
    args := c.RemainingArgs()
    filters := make([]*net.IPNet, 0)
    for _, arg := range args {
        _, cidr, err := net.ParseCIDR(arg)
        if (err != nil) {
            return plugin.Error("filterresponse", c.SyntaxErr(err.Error()))
        }
        filters = append(filters, cidr)
    }

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
        return FilterResponse{Next: next, Filters: filters}
	})

	return nil
}
