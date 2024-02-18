package rule

import (
	"fmt"
	"net"

	"github.com/paralleltree/mastoshield/lib"
)

type remoteIPAddressMatcher struct {
	targetRange *net.IPNet
}

func NewRemoteIPAddressMatcher(cidr string) (*remoteIPAddressMatcher, error) {
	_, targetRange, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("parse cidr: %w", err)
	}
	return &remoteIPAddressMatcher{
		targetRange: targetRange,
	}, nil
}

func (m *remoteIPAddressMatcher) Test(req *ProxyRequest) (bool, error) {
	remoteAddr, err := lib.ResolveClientIP(req.Request)
	if err != nil {
		return false, fmt.Errorf("resolve client addr: %w", err)
	}
	remoteIP := net.ParseIP(remoteAddr)
	if remoteIP == nil {
		return false, fmt.Errorf("cannot parse IP address: %s", remoteAddr)
	}
	return m.targetRange.Contains(remoteIP), nil
}
