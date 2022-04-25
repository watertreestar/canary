package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

// SetAlive for this backend
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

// IsAlive returns true when backend is alive
func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func (sp *ServerPool) GetNext() *Backend {
	next := sp.NextIndex()
	l := len(sp.backends) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(sp.backends) // take an index by modding with length
		// if we have an alive backend, use it and store if its not the original one
		if sp.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&sp.current, uint64(idx)) // mark the current one
			}
			return sp.backends[idx]
		}
	}
	return nil
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (sp *ServerPool) NextIndex() int {
	cur := sp.current
	return int(atomic.AddUint64(&cur, uint64(1)) % uint64(len(sp.backends)))
}

func (sp *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range sp.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

func (sp *ServerPool) HealthCheck() {
	for _, b := range sp.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

func (sp *ServerPool) healthCheck() {
	go func() {
		t := time.NewTicker(time.Minute * 2)
		for {
			select {
			case <-t.C:
				log.Println("Starting health check...")
				sp.HealthCheck()
				log.Println("Health check completed")
			}
		}
	}()
}

type Application struct {
	pool *ServerPool
}

func (app *Application) Init(server []string) {
	var pool ServerPool
	for _, address := range server {
		serverUrl, err := url.Parse(address)
		if err != nil {
			log.Fatal(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = ErrorHandler(pool, proxy)
		backend := Backend{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		}
		pool.AddBackend(&backend)
	}
	app.pool = &pool
}

func (app *Application) Start(port uint32) error {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(poolLoadBalance(*app.pool)),
	}
	app.pool.healthCheck()
	log.Printf("Load Balancer started at :%d\n", port)
	if err := server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
