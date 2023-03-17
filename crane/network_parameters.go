package crane

import "fmt"

type NetworkParameters struct {
	RawAlias       any    `json:"alias" yaml:"alias"`
	RawAliases     any    `json:"aliases" yaml:"aliases"`
	RawIp          string `json:"ip" yaml:"ip"`
	RawIpv4Address string `json:"ipv4_address" yaml:"ipv4_address"`
	RawIp6         string `json:"ip6" yaml:"ip6"`
	RawIpv6Address string `json:"ipv6_address" yaml:"ipv6_address"`
}

// If aliases are not defined in the config,
// the container name is added as a default alias.
// When an empty array is configured, no default alias is used.
func (n NetworkParameters) Alias(containerName string) []string {
	var aliases []string

	rawAliases := n.RawAliases
	if n.RawAlias != nil {
		rawAliases = n.RawAlias
	}

	if rawAliases == nil {
		aliases = append(aliases, containerName)
	} else {
		switch concreteValue := rawAliases.(type) {
		case []any:
			for _, v := range concreteValue {
				aliases = append(aliases, expandEnv(v.(string)))
			}
		}
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

func createNetworkParemetersFromMap(val map[string]any) NetworkParameters {
	var params NetworkParameters
	if val == nil {
		params = NetworkParameters{}
	} else {
		for k, v := range val {
			if k == "alias" || k == "aliases" {
				params.RawAlias = v.([]any)
			} else if k == "ip" || k == "ipv4_address" {
				params.RawIp = v.(string)
			} else if k == "ip6" || k == "ipv6_address" {
				params.RawIp6 = v.(string)
			} else {
				panic(StatusError{fmt.Errorf("unknown network key: %v", k), 65})
			}
		}
	}
	return params
}
