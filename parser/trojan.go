package parser

import (
	"fmt"
	"net/url"
	"strings"

	E "github.com/bestnite/sub2clash/error"
	P "github.com/bestnite/sub2clash/model/proxy"
)

type TrojanParser struct{}

func (p *TrojanParser) SupportClash() bool {
	return true
}

func (p *TrojanParser) SupportMeta() bool {
	return true
}

func (p *TrojanParser) GetPrefixes() []string {
	return []string{"trojan://"}
}

func (p *TrojanParser) GetType() string {
	return "trojan"
}

func (p *TrojanParser) Parse(proxy string) (P.Proxy, error) {
	if !hasPrefix(proxy, p.GetPrefixes()) {
		return P.Proxy{}, &E.ParseError{Type: E.ErrInvalidPrefix, Raw: proxy}
	}

	link, err := url.Parse(proxy)
	if err != nil {
		return P.Proxy{}, &E.ParseError{
			Type:    E.ErrInvalidStruct,
			Message: "url parse error",
			Raw:     proxy,
		}
	}

	password := link.User.Username()
	server := link.Hostname()
	if server == "" {
		return P.Proxy{}, &E.ParseError{
			Type:    E.ErrInvalidStruct,
			Message: "missing server host",
			Raw:     proxy,
		}
	}
	portStr := link.Port()
	if portStr == "" {
		return P.Proxy{}, &E.ParseError{
			Type:    E.ErrInvalidStruct,
			Message: "missing server port",
			Raw:     proxy,
		}
	}

	port, err := ParsePort(portStr)
	if err != nil {
		return P.Proxy{}, &E.ParseError{
			Type:    E.ErrInvalidPort,
			Message: err.Error(),
			Raw:     proxy,
		}
	}

	remarks := link.Fragment
	if remarks == "" {
		remarks = fmt.Sprintf("%s:%s", server, portStr)
	}
	remarks = strings.TrimSpace(remarks)

	query := link.Query()
	network, security, alpnStr, sni, pbk, sid, fp, path, host, serviceName := query.Get("type"), query.Get("security"), query.Get("alpn"), query.Get("sni"), query.Get("pbk"), query.Get("sid"), query.Get("fp"), query.Get("path"), query.Get("host"), query.Get("serviceName")

	var alpn []string
	if strings.Contains(alpnStr, ",") {
		alpn = strings.Split(alpnStr, ",")
	} else {
		alpn = nil
	}

	result := P.Trojan{
		Server:   server,
		Port:     port,
		Password: password,
		Network:  network,
	}

	if security == "xtls" || security == "tls" {
		result.Alpn = alpn
		result.Sni = sni
		result.TLS = true
	}

	if security == "reality" {
		result.TLS = true
		result.Sni = sni
		result.RealityOpts = P.RealityOptions{
			PublicKey: pbk,
			ShortID:   sid,
		}
		result.Fingerprint = fp
	}

	if network == "ws" {
		result.Network = "ws"
		result.WSOpts = P.WSOptions{
			Path: path,
			Headers: map[string]string{
				"Host": host,
			},
		}
	}

	if network == "grpc" {
		result.GrpcOpts = P.GrpcOptions{
			GrpcServiceName: serviceName,
		}
	}

	return P.Proxy{
		Type:   p.GetType(),
		Name:   remarks,
		Trojan: result,
	}, nil
}

func init() {
	RegisterParser(&TrojanParser{})
}
