package examples

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

// ConnectionPool manages a pool of ControlledClient connections.
// Clients are reused from an idle pool or created on demand up to maxSize.
// Stale connections (unused for longer than staleAfter) are discarded.
type ConnectionPool struct {
	uri        string
	password   string
	idle       chan *ControlledClient
	mu         sync.Mutex
	allocated  int
	maxSize    int
	staleAfter time.Duration
	logger     *log.Logger
}

// NewConnectionPool creates a new connection pool.
// maxSize is the maximum number of concurrent connections.
// staleAfter defines how long a connection can be idle before it's considered stale.
func NewConnectionPool(uri, password string, maxSize int, staleAfter time.Duration) *ConnectionPool {
	return &ConnectionPool{
		uri:        uri,
		password:   password,
		idle:       make(chan *ControlledClient, maxSize),
		maxSize:    maxSize,
		staleAfter: staleAfter,
		logger: log.New(
			log.Default().Writer(),
			"[RconPool] ",
			log.Default().Flags(),
		),
	}
}

// newClient creates and authenticates a new client.
func (p *ConnectionPool) newClient() (*ControlledClient, error) {
	client, err := NewControlledClient(p.uri)
	if err != nil {
		return nil, err
	}
	ok, err := client.Authenticate(p.password)
	if err != nil {
		client.Close()
		return nil, err
	}
	if !ok {
		client.Close()
		return nil, errors.New("authentication failed")
	}
	p.logger.Printf("new client created [allocated=%d]", p.allocated)
	return client, nil
}

// isStale checks if a client hasn't been used for longer than staleAfter.
func (p *ConnectionPool) isStale(client *ControlledClient) bool {
	return time.Now().Unix()-client.LastUsed() > int64(p.staleAfter.Seconds())
}

// Get acquires a client from the pool.
// Returns an idle client if available and not stale, or creates a new one if under capacity.
// If at capacity, blocks until a client is returned or ctx is cancelled.
func (p *ConnectionPool) Get(ctx context.Context) (*ControlledClient, error) {
	for {
		// Try non-blocking get from idle pool
		select {
		case client := <-p.idle:
			if p.isStale(client) {
				p.logger.Printf("discarding stale client [allocated=%d]", p.allocated)
				client.Close()
				p.mu.Lock()
				p.allocated--
				p.mu.Unlock()
				continue
			}
			return client, nil
		default:
		}

		// Try to allocate a new slot
		p.mu.Lock()
		if p.allocated < p.maxSize {
			p.allocated++
			p.mu.Unlock()
			client, err := p.newClient()
			if err != nil {
				p.mu.Lock()
				p.allocated--
				p.mu.Unlock()
				p.logger.Printf("client creation failed [allocated=%d]: %v", p.allocated, err)
				return nil, err
			}
			return client, nil
		}
		p.mu.Unlock()

		// At capacity, block until a client is available or ctx is done
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case client := <-p.idle:
			if p.isStale(client) {
				p.logger.Printf("discarding stale client [allocated=%d]", p.allocated)
				client.Close()
				p.mu.Lock()
				p.allocated--
				p.mu.Unlock()
				continue
			}
			return client, nil
		}
	}
}

// Release returns a client to the idle pool for reuse.
func (p *ConnectionPool) Release(client *ControlledClient) {
	select {
	case p.idle <- client:
	default:
		// Pool is full, close the excess client
		p.logger.Printf("pool full on release, closing excess client [allocated=%d]", p.allocated)
		client.Close()
		p.mu.Lock()
		p.allocated--
		p.mu.Unlock()
	}
}

// Discard marks a client as bad and removes it from the pool.
func (p *ConnectionPool) Discard(client *ControlledClient) {
	client.Close()
	p.mu.Lock()
	p.allocated--
	p.mu.Unlock()
}

// WithClient acquires a client, passes it to fn, and handles release/discard automatically.
// If fn returns an error, the client is discarded; otherwise it's released to the pool.
func (p *ConnectionPool) WithClient(ctx context.Context, fn func(*ControlledClient) error) error {
	client, err := p.Get(ctx)
	if err != nil {
		return err
	}
	if err := fn(client); err != nil {
		p.Discard(client)
		return err
	}
	p.Release(client)
	return nil
}

// Close closes all idle connections in the pool.
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		select {
		case client := <-p.idle:
			client.Close()
			p.allocated--
		default:
			p.logger.Printf("pool closed [remaining in-use=%d]", p.allocated)
			return
		}
	}
}
