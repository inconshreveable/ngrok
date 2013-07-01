package server

import (
	"fmt"
	"net"
	"ngrok/cache"
	"sync"
	"time"
)

const (
	cacheSaveInterval time.Duration = 10 * time.Minute
)

type cacheUrl string

func (url cacheUrl) Size() int {
	return len(url)
}

// TunnelRegistry maps a tunnel URL to Tunnel structures
type TunnelRegistry struct {
	tunnels  map[string]*Tunnel
	affinity *cache.LRUCache
	sync.RWMutex
}

func NewTunnelRegistry(cacheSize uint64, cacheFile string) *TunnelRegistry {
	manager := &TunnelRegistry{
		tunnels:  make(map[string]*Tunnel),
		affinity: cache.NewLRUCache(cacheSize),
	}

	if cacheFile != "" {
		// load cache entries from file
		manager.affinity.LoadItemsFromFile(cacheFile)

		// save cache periodically to file
		manager.SaveCacheThread(cacheFile, cacheSaveInterval)
	}

	return manager
}

// Spawns a goroutine the periodically saves the cache to a file.
func (r *TunnelRegistry) SaveCacheThread(path string, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			r.affinity.SaveItemsToFile(path)
		}
	}()
}

// Register a tunnel with a specific url, returns an error
// if a tunnel is already registered at that url
func (r *TunnelRegistry) Register(url string, t *Tunnel) error {
	r.Lock()
	defer r.Unlock()

	if r.tunnels[url] != nil {
		return fmt.Errorf("The tunnel %s is already registered.", url)
	}

	r.tunnels[url] = t

	return nil
}

// Register a tunnel with the following process:
// Consult the affinity cache to try to assign a previously used tunnel url if possible
// Generate new urls repeatedly with the urlFn and register until one is available.
func (r *TunnelRegistry) RegisterRepeat(urlFn func() string, t *Tunnel) string {
	var url string

	clientIp := t.ctl.conn.RemoteAddr().(*net.TCPAddr).IP.String()
	clientId := t.regMsg.ClientId

	ipCacheKey := fmt.Sprintf("client-ip:%s", clientIp)
	idCacheKey := fmt.Sprintf("client-id:%s", clientId)

	// check cache for ID first, because we prefer that over IP which might
	// not be specific to a user because of NATs
	if v, ok := r.affinity.Get(idCacheKey); ok {
		url = string(v.(cacheUrl))
		t.Debug("Found registry affinity %s for %s", url, idCacheKey)
	} else if v, ok := r.affinity.Get(ipCacheKey); ok {
		url = string(v.(cacheUrl))
		t.Debug("Found registry affinity %s for %s", url, ipCacheKey)
	} else {
		url = urlFn()
	}

	for {
		if err := r.Register(url, t); err != nil {
			// pick a new url and try again
			url = urlFn()
		} else {
			// we successfully assigned a url, we're done

			// save our choice in the cache
			r.affinity.Set(ipCacheKey, cacheUrl(url))
			r.affinity.Set(idCacheKey, cacheUrl(url))

			return url
		}
	}
}

func (r *TunnelRegistry) Del(url string) {
	r.Lock()
	defer r.Unlock()
	delete(r.tunnels, url)
}

func (r *TunnelRegistry) Get(url string) *Tunnel {
	r.RLock()
	defer r.RUnlock()
	return r.tunnels[url]
}
