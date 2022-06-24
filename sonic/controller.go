package sonic

// controller defines base interface around driversPool.
type controller struct {
	*driversPool
}

func newController(
	opts controllerOptions,
) (*controller, error) {
	df := driverFactory{
		Host:     opts.Host,
		Port:     opts.Port,
		Password: opts.Password,
		Channel:  opts.Channel,
	}

	dp, err := newDriversPool(
		&df,
		opts.PoolMinConnections,
		opts.PoolMaxConnections,
		opts.PoolPingThreshold,
	)
	if err != nil {
		return nil, err
	}

	return &controller{
		driversPool: dp,
	}, nil
}

// Quit all connections and close the pool. It never returns an error.
func (c *controller) Quit() error {
	c.driversPool.Close()

	return nil
}

// Ping one connection.
func (c *controller) Ping() error {
	d, err := c.Get()
	if err != nil {
		return err
	}

	return d.Ping()
}
