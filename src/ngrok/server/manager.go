package server

import (
	"fmt"
	cache "github.com/pmylund/go-cache"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	cacheDuration        time.Duration = 24 * time.Hour
	cacheCleanupInterval time.Duration = time.Minute
)

/**
 * TunnelManager: Manages a set of tunnels
 */
type TunnelManager struct {
	domain           string
	tunnels          map[string]*Tunnel
	idDomainAffinity *cache.Cache
	ipDomainAffinity *cache.Cache
	sync.RWMutex
}

func NewTunnelManager(domain string) *TunnelManager {
	return &TunnelManager{
		domain:           domain,
		tunnels:          make(map[string]*Tunnel),
		idDomainAffinity: cache.New(cacheDuration, cacheCleanupInterval),
		ipDomainAffinity: cache.New(cacheDuration, cacheCleanupInterval),
	}
}

func (m *TunnelManager) Add(t *Tunnel) error {
	assignTunnel := func(url string) bool {
		m.Lock()
		defer m.Unlock()

		if m.tunnels[url] == nil {
			m.tunnels[url] = t
			return true
		}

		return false
	}

	url := ""
	switch t.regMsg.Protocol {
	case "tcp":
		addr := t.listener.Addr().(*net.TCPAddr)
		url = fmt.Sprintf("tcp://%s:%d", m.domain, addr.Port)
		if !assignTunnel(url) {
			return t.Error("TCP at %s already registered!", url)
		}

	case "http":
		if strings.TrimSpace(t.regMsg.Hostname) != "" {
			url = fmt.Sprintf("http://%s", t.regMsg.Hostname)
		} else if strings.TrimSpace(t.regMsg.Subdomain) != "" {
			url = fmt.Sprintf("http://%s.%s", t.regMsg.Subdomain, m.domain)
		}

		if url != "" {
			if !assignTunnel(url) {
				return t.Warn("The tunnel address %s is already registered!", url)
			}
		} else {
			clientIp := t.ctl.conn.RemoteAddr().(*net.TCPAddr).IP.String()
			clientId := t.regMsg.ClientId

			// try to give the same subdomain back if it's available
			subdomain := fmt.Sprintf("%x", rand.Int31())
			if lastDomain, ok := m.idDomainAffinity.Get(clientId); ok {
				t.Debug("Found affinity for subdomain %s with client id %s", subdomain, clientId)
				subdomain = lastDomain.(string)
			} else if lastDomain, ok = m.ipDomainAffinity.Get(clientIp); ok {
				t.Debug("Found affinity for subdomain %s with client ip %s", subdomain, clientIp)
				subdomain = lastDomain.(string)
			}

			// pick one randomly
			for {
				url = fmt.Sprintf("http://%s.%s", subdomain, m.domain)
				if assignTunnel(url) {
					break
				} else {
					subdomain = fmt.Sprintf("%x", rand.Int31())
				}
			}

			// save our choice so we can try to give clients back the same
			// tunnel later
			m.idDomainAffinity.Set(clientId, subdomain, 0)
			m.ipDomainAffinity.Set(clientIp, subdomain, 0)
		}

	default:
		return t.Error("Unrecognized protocol type %s", t.regMsg.Protocol)
	}

	t.url = url

	return nil
}

func (m *TunnelManager) Del(url string) {
	m.Lock()
	defer m.Unlock()
	delete(m.tunnels, url)
}

func (m *TunnelManager) Get(url string) *Tunnel {
	m.RLock()
	defer m.RUnlock()
	return m.tunnels[url]
}
