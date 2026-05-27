package scrapper

import (
	"sync"
)

// ProxyPool manages a thread-safe list of proxies to iterate over.
type ProxyPool struct {
	Proxies []string
	Current int
	Mu      sync.Mutex
}

// GlobalProxyPool holds the initialised proxies for the application.
var GlobalProxyPool *ProxyPool = &ProxyPool{
	Proxies: []string{}, 
}

// InitProxyPool initializes the global pool with a list of proxy strings.
// A proxy str looks like: "http://username:password@ip:port" or "http://ip:port".
func InitProxyPool(proxies []string) {
	GlobalProxyPool.Mu.Lock()
	defer GlobalProxyPool.Mu.Unlock()
	
	GlobalProxyPool.Proxies = proxies
	GlobalProxyPool.Current = 0
}

// GetNext returns the next proxy URL in the pool using a Round-Robin strategy.
// If the pool is empty, it returns an empty string (signaling a direct connection).
func (p *ProxyPool) GetNext() string {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if len(p.Proxies) == 0 {
		return "" // fallback to no proxy
	}

	proxy := p.Proxies[p.Current]
	p.Current = (p.Current + 1) % len(p.Proxies)
	return proxy
}
