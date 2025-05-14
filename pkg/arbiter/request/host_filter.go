package request

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type HostFilter struct {
	blockedCIDRs []*net.IPNet
	allowedCIDRs []*net.IPNet
	blockedHosts map[string]bool
	allowedHosts map[string]bool
	dnsCache     *TTLCache
	mu           sync.RWMutex
}

type DNSRecord struct {
	key       string
	ips       []net.IP
	expiresAt time.Time
}

type TTLCache struct {
	cache    map[string]*list.Element
	lruList  *list.List
	capacity int
	ttl      time.Duration
	mu       sync.RWMutex
}

func NewTTLCache(capacity int, ttl time.Duration) *TTLCache {
	return &TTLCache{
		cache:    make(map[string]*list.Element),
		lruList:  list.New(),
		capacity: capacity,
		ttl:      ttl,
	}
}

func (c *TTLCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	element, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	record := element.Value.(*DNSRecord)
	if time.Now().After(record.expiresAt) {
		return nil, false
	}

	c.lruList.MoveToFront(element)
	return record.ips, true
}

func (c *TTLCache) Set(key string, value []net.IP) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cache) >= c.capacity {
		c.removeOldest()
	}

	record := &DNSRecord{
		key:       key,
		ips:       value,
		expiresAt: time.Now().Add(c.ttl),
	}

	element := c.lruList.PushFront(record)
	c.cache[key] = element
}

func (c *TTLCache) removeOldest() {
	element := c.lruList.Back()
	if element != nil {
		c.lruList.Remove(element)
		delete(c.cache, element.Value.(*DNSRecord).key)
	}
}

func PrivateCIDRs() []string {
	return []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
	}
}

func NewHostFilter(blockedCIDRs, allowedCIDRs, blockedHosts, allowedHosts []string, cacheCap int, cacheTTL time.Duration) *HostFilter {
	blockedH := map[string]bool{}
	for _, host := range blockedHosts {
		blockedH[host] = true
	}
	allowedH := map[string]bool{}
	for _, host := range allowedHosts {
		allowedH[host] = true
	}

	return &HostFilter{
		blockedCIDRs: parseCIDRs(blockedCIDRs),
		allowedCIDRs: parseCIDRs(allowedCIDRs),
		blockedHosts: blockedH,
		allowedHosts: allowedH,
		dnsCache:     NewTTLCache(cacheCap, cacheTTL)}
}

func (f *HostFilter) AddBlockedCIDR(cidrs []string) {
	f.blockedCIDRs = append(f.blockedCIDRs, parseCIDRs(cidrs)...)
}

func (f *HostFilter) AddAllowedCIDR(cidrs []string) {
	f.allowedCIDRs = append(f.allowedCIDRs, parseCIDRs(cidrs)...)
}

func (f *HostFilter) AddBlockedHost(hosts []string) {
	for _, h := range hosts {
		f.blockedHosts[strings.ToLower(h)] = true
	}
}

func (f *HostFilter) AddAllowedHost(hosts []string) {
	for _, h := range hosts {
		f.allowedHosts[strings.ToLower(h)] = true
	}
}

func (f *HostFilter) IsBlocked(host string) (bool, error) {
	host = strings.ToLower(host)

	if f.blockedHosts[host] {
		return true, errors.New("host is blocked")
	}

	if len(f.allowedHosts) > 0 && !f.allowedHosts[host] {
		return true, errors.New("host is not in allowed list")
	}

	ips, err := f.lookupIP(host)
	if err != nil {
		return true, fmt.Errorf("failed to resolve host: %v", err)
	}

	for _, ip := range ips {
		if f.isIPBlocked(ip) {
			return true, fmt.Errorf("resolved IP %s is blocked", ip)
		}
	}

	return false, nil
}

func (f *HostFilter) isIPBlocked(ip net.IP) bool {
	for _, cidr := range f.blockedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	if len(f.allowedCIDRs) > 0 {
		for _, cidr := range f.allowedCIDRs {
			if cidr.Contains(ip) {
				return false
			}
		}
		return true
	}

	return false
}

func (f *HostFilter) lookupIP(host string) ([]net.IP, error) {
	f.mu.RLock()
	if cached, ok := f.dnsCache.Get(host); ok {
		f.mu.RUnlock()
		return cached.([]net.IP), nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if cached, ok := f.dnsCache.Get(host); ok {
		return cached.([]net.IP), nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	f.dnsCache.Set(host, ips)
	return ips, nil
}

func parseCIDRs(cidrs []string) []*net.IPNet {
	var result []*net.IPNet
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			result = append(result, ipNet)
		}
	}
	return result
}

type FilteringDialer struct {
	hostFilter *HostFilter
	dialer     *net.Dialer
}

func NewFilteringDialer(filter *HostFilter) *FilteringDialer {
	return &FilteringDialer{
		hostFilter: filter,
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}
}

func (d *FilteringDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	if blocked, err := d.hostFilter.IsBlocked(host); blocked {
		return nil, err
	}

	return d.dialer.DialContext(ctx, network, addr)
}
