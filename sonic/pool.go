package sonic

import (
	"sync"
	"time"
)

type driversPool struct {
	driverFactory *driverFactory
	drivers       chan *driverWrapper
	pingThreshold time.Duration

	isPoolClosedMu sync.RWMutex
	isPoolClosed   bool
}

func newDriversPool(
	df *driverFactory,
	minIdle int,
	maxIdle int,
	pingThreshold time.Duration,
) (*driversPool, error) {
	dp := &driversPool{
		driverFactory: df,
		drivers:       make(chan *driverWrapper, maxIdle),

		pingThreshold: pingThreshold,

		isPoolClosedMu: sync.RWMutex{},
		isPoolClosed:   false,
	}

	var err error
	var dw *driverWrapper

	// Open connnections.
	drivers := make([]*driverWrapper, 0, minIdle)
	for i := 0; i < maxIdle; i++ {
		dw, err = dp.Get()
		if err != nil {
			// We still need to close already opened connections.
			break
		}

		drivers = append(drivers, dw)
	}

	// Return all connections to the pool.
	for _, d := range drivers {
		d.close()
	}

	return dp, err
}

// put the connection back.
func (p *driversPool) put(dw *driverWrapper) {
	if dw.closed {
		return
	}

	p.isPoolClosedMu.RLock()
	defer p.isPoolClosedMu.RUnlock()

	if p.isPoolClosed {
		dw.close()

		return
	}

	select {
	case p.drivers <- dw:
	default:
		// The pool is full.
		_ = dw.Quit()
		dw.close()
	}
}

// Get a healthy driver from the pool. It pings the connection
// if it was configured by OptionPoolPingThreshold.
// It will open a connection if no connection is available in the pool.
// Closing of connection will return it back.
func (p *driversPool) Get() (*driverWrapper, error) {
	p.isPoolClosedMu.RLock()
	defer p.isPoolClosedMu.RUnlock()

	if p.isPoolClosed {
		return nil, ErrClosed
	}

	select {
	case d := <-p.drivers:
		if !d.softPing(p.pingThreshold) {
			d.close()

			return p.Get()
		}

		return d, nil
	default:
		d := p.driverFactory.Build()

		if err := d.Connect(); err != nil {
			return nil, err
		}

		return p.wrapDriver(d), nil
	}
}

// Close and quit all connections in the pool.
func (p *driversPool) Close() {
	p.isPoolClosedMu.Lock()
	defer p.isPoolClosedMu.Unlock()
	p.isPoolClosed = true

	close(p.drivers)
	for dw := range p.drivers {
		if !dw.closed {
			_ = dw.Quit()

			dw.close()
		}
	}
}

// wrapDriver overrides driver's connection close method.
func (p *driversPool) wrapDriver(d *driver) *driverWrapper {
	return &driverWrapper{
		driver:  d,
		onClose: p.put,
	}
}

// driverWrapper helps to override close for *driver.connection.
type driverWrapper struct {
	onClose func(*driverWrapper)
	*driver
}

// close overrides close method of the driver.
func (dw *driverWrapper) close() {
	dw.onClose(dw)
}
