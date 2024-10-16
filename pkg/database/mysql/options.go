package mysql

import (
	"time"
)

// Option - Option type
type Option func(*Mysql)

// MaxOpenConnections - Defines maximum open connections for mysql db
func MaxOpenConnections(size int) Option {
	return func(c *Mysql) {
		c.maxOpenConnections = size
	}
}

// MaxIdleConnections - Defines maximum idle connections for mysql db
func MaxIdleConnections(c int) Option {
	return func(p *Mysql) {
		p.maxIdleConnections = c
	}
}

// MaxConnectionIdleTime - Defines maximum connection idle for mysql db
func MaxConnectionIdleTime(d time.Duration) Option {
	return func(p *Mysql) {
		p.maxConnectionIdleTime = d
	}
}

// MaxConnectionLifeTime - Defines maximum connection lifetime for mysql db
func MaxConnectionLifeTime(d time.Duration) Option {
	return func(p *Mysql) {
		p.maxConnectionLifeTime = d
	}
}

func MaxDataPerWrite(v int) Option {
	return func(c *Mysql) {
		c.maxDataPerWrite = v
	}
}

func WatchBufferSize(v int) Option {
	return func(c *Mysql) {
		c.watchBufferSize = v
	}
}

func MaxRetries(v int) Option {
	return func(c *Mysql) {
		c.maxRetries = v
	}
}
