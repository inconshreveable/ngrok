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

/**
 * TunnelManager: Manages a set of tunnels
 */
type TunnelManager struct {
	domain         string
	tunnels        map[string]*Tunnel
	domainAffinity *cache.Cache
	sync.RWMutex
}

func NewTunnelManager(domain string) *TunnelManager {
	return &TunnelManager{
		domain:         domain,
		tunnels:        make(map[string]*Tunnel),
		domainAffinity: cache.New(24*time.Hour, time.Minute),
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
		metrics.tcpTunnelMeter.Mark(1)

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
			// try to give the same subdomain back if it's available
			subdomain, ok := m.domainAffinity.Get(t.regMsg.ClientId)
			if !ok {
				subdomain = fmt.Sprintf("%x", rand.Int31())
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

			// save our choice for later
			// XXX: this is going to leak memory
			m.domainAffinity.Set(t.regMsg.ClientId, subdomain, 0)
		}

	default:
		return t.Error("Unrecognized protocol type %s", t.regMsg.Protocol)
	}

	t.url = url
	metrics.tunnelMeter.Mark(1)
	//metrics.tunnelGauge.Update(int64(len(m.tunnels)))

	switch t.regMsg.OS {
	case "windows":
		metrics.windowsCounter.Inc(1)
	case "linux":
		metrics.linuxCounter.Inc(1)
	case "darwin":
		metrics.osxCounter.Inc(1)
	default:
		metrics.otherCounter.Inc(1)
	}

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
