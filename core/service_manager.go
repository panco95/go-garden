package core

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/panco95/go-garden/core/log"
	clientV3 "go.etcd.io/etcd/client/v3"
)

// node is one registered instance. Addr is ip:httpPort:rpcPort.
// Waiting and Finish are updated atomically by in-flight calls and must be
// read with atomic loads. The pointer stays valid after the node leaves the
// registry, so a call that already selected it can finish safely.
type node struct {
	Addr    string
	Waiting int64
	Finish  int64
}

type service struct {
	Nodes map[string]*node
}

type routedNode struct {
	serviceName string
	node        *node
}

// GetServices returns a snapshot of the cluster. The snapshot does not share
// node pointers with the live registry.
func (g *Garden) GetServices() map[string]*service {
	g.servicesMu.RLock()
	defer g.servicesMu.RUnlock()
	out := make(map[string]*service, len(g.services))
	for name, svc := range g.services {
		nodes := make(map[string]*node, len(svc.Nodes))
		for addr, n := range svc.Nodes {
			copied := *n
			nodes[addr] = &copied
		}
		out[name] = &service{Nodes: nodes}
	}
	return out
}

// GetServiceIp get this service ip
func (g *Garden) GetServiceIp() string {
	return g.cfg.Service.ServiceIp
}

func (g *Garden) localNodeAddr() string {
	return g.cfg.Service.ServiceIp + ":" + g.cfg.Service.HttpPort + ":" + g.cfg.Service.RpcPort
}

// GetServiceId get this service union id
func (g *Garden) GetServiceId() string {
	return g.cfg.Service.EtcdKey + "_" + g.cfg.Service.ServiceName + "_" + g.localNodeAddr()
}

func (g *Garden) servicePrefix() string {
	return g.cfg.Service.EtcdKey + "_"
}

func (g *Garden) bootService() error {
	g.services = map[string]*service{}
	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	g.life.mu.Lock()
	g.life.syncCancel = cancel
	g.life.mu.Unlock()

	if g.cfg.Service.ServiceIp == "" {
		ip, err := getOutboundIP()
		if err != nil {
			cancel()
			return err
		}
		g.cfg.Service.ServiceIp = ip
	}

	if err := g.serviceRegister(ctx, true); err != nil {
		cancel()
		return err
	}
	return nil
}

func (g *Garden) serviceRegister(ctx context.Context, startWatch bool) error {
	if g.isShuttingDown() || ctx.Err() != nil {
		return errors.New("shutting down")
	}
	client, err := g.GetEtcd()
	if err != nil {
		return err
	}
	// New lease
	resp, err := client.Grant(context.TODO(), 2)
	if err != nil {
		return err
	}
	g.setLease(int64(resp.ID))
	_, err = client.Put(context.TODO(), g.GetServiceId(), "0", clientV3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	// keep alive
	ch, err := client.KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		return err
	}
	// monitor etcd connection
	go func() {
		for resp := range ch {
			if resp == nil {
				break
			}
		}
		if g.isShuttingDown() || ctx.Err() != nil {
			return
		}
		if err := g.serviceRegister(ctx, false); err != nil && !g.isShuttingDown() {
			log.Error("serviceRegister", err)
		}
	}()

	if startWatch {
		go g.serviceSyncLoop(ctx)
	}
	return nil
}

// serviceSyncLoop keeps the local registry aligned with etcd.
// Each pass lists the prefix and adds or deletes nodes to match that snapshot,
// then watches from the next revision. When the watch closes, the loop starts
// again with a fresh list so a dropped watch cannot leave stale nodes behind.
func (g *Garden) serviceSyncLoop(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		rev, err := g.syncServicesFromEtcd()
		if err != nil {
			log.Error("syncServices", err)
			if !waitCtx(ctx, time.Second) {
				return
			}
			continue
		}
		err = g.watchServices(ctx, rev)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Error("watchServices", err)
			if !waitCtx(ctx, time.Second) {
				return
			}
		}
	}
}

func waitCtx(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (g *Garden) syncServicesFromEtcd() (int64, error) {
	client, err := g.GetEtcd()
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := client.Get(ctx, g.servicePrefix(), clientV3.WithPrefix())
	if err != nil {
		return 0, err
	}

	desired := make(map[string]map[string]struct{})
	for _, kv := range resp.Kvs {
		name, addr, err := g.parseServiceKey(string(kv.Key))
		if err != nil {
			log.Error("parseServiceKey", err)
			continue
		}
		if desired[name] == nil {
			desired[name] = map[string]struct{}{}
		}
		desired[name][addr] = struct{}{}
	}

	g.servicesMu.Lock()
	joined, left := g.reconcileServices(desired)
	var rev int64
	if resp.Header != nil {
		rev = resp.Header.Revision
	}
	g.servicesMu.Unlock()

	for _, item := range joined {
		log.Infof("service", "%s node %s join", item.serviceName, item.node.Addr)
	}
	for _, item := range left {
		log.Infof("service", "%s node %s leave", item.serviceName, item.node.Addr)
	}
	g.closeRPCClients(left)
	return rev, nil
}

func (g *Garden) watchServices(ctx context.Context, syncedRev int64) error {
	client, err := g.GetEtcd()
	if err != nil {
		return err
	}
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []clientV3.OpOption{clientV3.WithPrefix()}
	if syncedRev > 0 {
		opts = append(opts, clientV3.WithRev(syncedRev+1))
	}
	rch := client.Watch(watchCtx, g.servicePrefix(), opts...)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case wresp, ok := <-rch:
			if !ok {
				return errors.New("service watch closed")
			}
			if err := wresp.Err(); err != nil {
				return err
			}
			g.applyWatchEvents(wresp.Events)
		}
	}
}

func (g *Garden) applyWatchEvents(events []*clientV3.Event) {
	g.servicesMu.Lock()
	var joined, left []routedNode
	for _, ev := range events {
		if ev.Kv == nil {
			continue
		}
		name, addr, err := g.parseServiceKey(string(ev.Kv.Key))
		if err != nil {
			log.Error("parseServiceKey", err)
			continue
		}
		switch ev.Type {
		case clientV3.EventTypePut:
			if g.addNodeLocked(name, addr) {
				joined = append(joined, routedNode{serviceName: name, node: &node{Addr: addr}})
			}
		case clientV3.EventTypeDelete:
			if g.delNodeLocked(name, addr) {
				left = append(left, routedNode{serviceName: name, node: &node{Addr: addr}})
			}
		}
	}
	g.servicesMu.Unlock()

	for _, item := range joined {
		log.Infof("service", "%s node %s join", item.serviceName, item.node.Addr)
	}
	for _, item := range left {
		log.Infof("service", "%s node %s leave", item.serviceName, item.node.Addr)
	}
	g.closeRPCClients(left)
}

// parseServiceKey splits {etcdKey}_{serviceName}_{ip}:{httpPort}:{rpcPort}.
// The address is the segment after the last underscore, so a service name may contain underscores.
func (g *Garden) parseServiceKey(key string) (string, string, error) {
	prefix := g.servicePrefix()
	if !strings.HasPrefix(key, prefix) {
		return "", "", errors.New("invalid service key")
	}
	rest := strings.TrimPrefix(key, prefix)
	idx := strings.LastIndex(rest, "_")
	if idx <= 0 || idx == len(rest)-1 {
		return "", "", errors.New("invalid service key")
	}
	return rest[:idx], rest[idx+1:], nil
}

func (g *Garden) getServicesByName(serviceName string) ([]string, error) {
	client, err := g.GetEtcd()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.Get(ctx, g.servicePrefix()+serviceName+"_", clientV3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var addrs []string
	for _, ev := range resp.Kvs {
		name, addr, err := g.parseServiceKey(string(ev.Key))
		if err != nil || name != serviceName {
			continue
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

// reconcileServices makes the local registry match desired.
// Existing node pointers are kept so in-flight Waiting counts survive a resync.
// Caller must hold servicesMu.
func (g *Garden) reconcileServices(desired map[string]map[string]struct{}) (joined, left []routedNode) {
	for name, svc := range g.services {
		addrs, ok := desired[name]
		if !ok {
			for addr := range svc.Nodes {
				left = append(left, routedNode{serviceName: name, node: &node{Addr: addr}})
			}
			delete(g.services, name)
			continue
		}
		for addr := range svc.Nodes {
			if _, keep := addrs[addr]; !keep {
				delete(svc.Nodes, addr)
				left = append(left, routedNode{serviceName: name, node: &node{Addr: addr}})
			}
		}
		if len(svc.Nodes) == 0 {
			delete(g.services, name)
		}
	}
	for name, addrs := range desired {
		if _, ok := g.services[name]; !ok {
			g.services[name] = &service{Nodes: map[string]*node{}}
		}
		for addr := range addrs {
			if g.addNodeLocked(name, addr) {
				joined = append(joined, routedNode{serviceName: name, node: &node{Addr: addr}})
			}
		}
	}
	return joined, left
}

// addNodeLocked inserts a node once. Caller must hold servicesMu.
func (g *Garden) addNodeLocked(name, addr string) bool {
	svc, ok := g.services[name]
	if !ok {
		svc = &service{Nodes: map[string]*node{}}
		g.services[name] = svc
	}
	if _, exists := svc.Nodes[addr]; exists {
		return false
	}
	svc.Nodes[addr] = &node{Addr: addr}
	return true
}

// delNodeLocked removes a node by address. Caller must hold servicesMu.
func (g *Garden) delNodeLocked(name, addr string) bool {
	svc, ok := g.services[name]
	if !ok {
		return false
	}
	if _, exists := svc.Nodes[addr]; !exists {
		return false
	}
	delete(svc.Nodes, addr)
	if len(svc.Nodes) == 0 {
		delete(g.services, name)
	}
	return true
}

func (g *Garden) listServiceNodes() []routedNode {
	g.servicesMu.RLock()
	defer g.servicesMu.RUnlock()
	list := make([]routedNode, 0)
	for name, svc := range g.services {
		for _, n := range svc.Nodes {
			list = append(list, routedNode{serviceName: name, node: n})
		}
	}
	return list
}

// nodeHTTPAddr and nodeRPCAddr split ip:httpPort:rpcPort.
func nodeHTTPAddr(addr string) (string, error) {
	ip, httpPort, _, err := splitNodeAddr(addr)
	if err != nil {
		return "", err
	}
	return ip + ":" + httpPort, nil
}

func nodeRPCAddr(addr string) (string, error) {
	ip, _, rpcPort, err := splitNodeAddr(addr)
	if err != nil {
		return "", err
	}
	return ip + ":" + rpcPort, nil
}

func splitNodeAddr(addr string) (ip, httpPort, rpcPort string, err error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", errors.New("invalid node address")
	}
	return parts[0], parts[1], parts[2], nil
}

// selectService copies the live node pointers, then picks the smallest Waiting.
// Ties are broken at random among the nodes that share that waiting count.
func (g *Garden) selectService(name string) (*node, error) {
	return g.selectServiceExcept(name, nil)
}

// selectServiceExcept copies live node pointers except skip, then picks the smallest Waiting.
func (g *Garden) selectServiceExcept(name string, skip map[string]struct{}) (*node, error) {
	g.servicesMu.RLock()
	svc, ok := g.services[name]
	if !ok {
		g.servicesMu.RUnlock()
		return nil, errors.New("service not found")
	}
	nodes := make([]*node, 0, len(svc.Nodes))
	for _, n := range svc.Nodes {
		if _, excluded := skip[n.Addr]; excluded {
			continue
		}
		nodes = append(nodes, n)
	}
	g.servicesMu.RUnlock()
	if len(nodes) == 0 {
		return nil, errors.New("service node not found")
	}
	return pickLeastWaiting(nodes), nil
}

func pickLeastWaiting(nodes []*node) *node {
	best := atomic.LoadInt64(&nodes[0].Waiting)
	chosen := []*node{nodes[0]}
	for _, n := range nodes[1:] {
		waiting := atomic.LoadInt64(&n.Waiting)
		if waiting < best {
			best = waiting
			chosen = []*node{n}
		} else if waiting == best {
			chosen = append(chosen, n)
		}
	}
	if len(chosen) == 1 {
		return chosen[0]
	}
	return chosen[rand.Intn(len(chosen))]
}
