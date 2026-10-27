package core

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panco95/go-garden/core/log"
	rpcx_logger "github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/server"
	clientV3 "go.etcd.io/etcd/client/v3"
)

type lifecycle struct {
	mu           sync.Mutex
	leaseID      int64
	syncCancel   context.CancelFunc
	tracerCloser io.Closer
	shuttingDown int32
}

// Run starts HTTP and RPC, then blocks until SIGINT or SIGTERM.
// It removes this process from etcd and stops both servers before returning.
func (g *Garden) Run(route func(r *gin.Engine), rpc interface{}, auth func() gin.HandlerFunc) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return g.serve(ctx, route, rpc, auth)
}

func (g *Garden) serve(ctx context.Context, route func(r *gin.Engine), rpc interface{}, auth func() gin.HandlerFunc) error {
	httpSrv, err := g.startHTTP(route, auth)
	if err != nil {
		return err
	}
	rpcSrv, err := g.startRPC(rpc)
	if err != nil {
		g.shutdown(httpSrv, nil)
		return err
	}
	<-ctx.Done()
	log.Info("shutdown", "stopping")
	g.shutdown(httpSrv, rpcSrv)
	return nil
}

func (g *Garden) startHTTP(route func(r *gin.Engine), auth func() gin.HandlerFunc) (*http.Server, error) {
	engine, err := g.httpEngine(route, auth)
	if err != nil {
		return nil, err
	}
	address := g.GetServiceIp()
	if g.cfg.Service.HttpOut {
		address = "0.0.0.0"
	}
	srv := &http.Server{
		Addr:    address + ":" + g.cfg.Service.HttpPort,
		Handler: engine,
	}
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return nil, err
	}
	log.Infof("http", "listen on: %s", srv.Addr)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Error("gin", err)
		}
	}()
	return srv, nil
}

func (g *Garden) startRPC(rpc interface{}) (*server.Server, error) {
	address := g.GetServiceIp()
	if g.cfg.Service.RpcOut {
		address = "0.0.0.0"
	}
	addr := address + ":" + g.cfg.Service.RpcPort
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := server.NewServer()
	s.AuthFunc = g.rpcAuth
	if logger := log.GetLogger(); logger != nil {
		rpcx_logger.SetLogger(logger)
	}
	if err := s.RegisterName(g.cfg.Service.ServiceName, rpc, ""); err != nil {
		ln.Close()
		return nil, err
	}
	log.Infof("rpc", "listen on: %s", addr)
	go func() {
		if err := s.ServeListener("tcp", ln); err != nil {
			log.Error("rpcRun", err)
		}
	}()
	return s, nil
}

func (g *Garden) shutdown(httpSrv *http.Server, rpcSrv *server.Server) {
	atomic.StoreInt32(&g.life.shuttingDown, 1)
	g.stopSync()
	g.deregister()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if httpSrv != nil {
		if err := httpSrv.Shutdown(ctx); err != nil {
			log.Error("http shutdown", err)
		}
	}
	if rpcSrv != nil {
		if err := rpcSrv.Shutdown(ctx); err != nil {
			log.Error("rpc shutdown", err)
		}
	}
	g.closeAllRPCClients()
	g.closeTracer()
}

func (g *Garden) isShuttingDown() bool {
	return atomic.LoadInt32(&g.life.shuttingDown) == 1
}

func (g *Garden) stopSync() {
	g.life.mu.Lock()
	cancel := g.life.syncCancel
	g.life.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (g *Garden) setLease(id int64) {
	g.life.mu.Lock()
	g.life.leaseID = id
	g.life.mu.Unlock()
}

func (g *Garden) deregister() {
	client, err := g.GetEtcd()
	if err != nil {
		return
	}
	g.life.mu.Lock()
	id := g.life.leaseID
	g.life.mu.Unlock()
	if id == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := client.Revoke(ctx, clientV3.LeaseID(id)); err != nil {
		log.Error("deregister", err)
	}
}

func (g *Garden) closeTracer() {
	g.life.mu.Lock()
	closer := g.life.tracerCloser
	g.life.tracerCloser = nil
	g.life.mu.Unlock()
	if closer != nil {
		closer.Close()
	}
}
