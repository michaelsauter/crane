package crane

type NetworkParameters struct {
	RawAlias       []string `json:"alias" yaml:"alias"`
	RawAliases     []string `json:"aliases" yaml:"aliases"`
	RawIp          string   `json:"ip" yaml:"ip"`
	RawIpv4Address string   `json:"ipv4_address" yaml:"ipv4_address"`
	RawIp6         string   `json:"ip6" yaml:"ip6"`
	RawIpv6Address string   `json:"ipv6_address" yaml:"ipv6_address"`
}

func (n NetworkParameters) Alias() []string {
	var aliases []string
	rawAliases := n.RawAliases
	if len(n.RawAlias) > 0 {
		rawAliases = n.RawAlias
	}
	for _, rawAlias := range rawAliases {
		aliases = append(aliases, expandEnv(rawAlias))
	}
	return aliases
}

func (n NetworkParameters) Ip() string {
	if len(n.RawIp) > 0 {
		return expandEnv(n.RawIp)
	}
	return expandEnv(n.RawIpv4Address)
}

func (n NetworkParameters) Ip6() string {
	if len(n.RawIp6) > 0 {
		return expandEnv(n.RawIp6)
	}
	return expandEnv(n.RawIpv6Address)
}
