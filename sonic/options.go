package sonic

import "time"

type controllerOptions struct {
	Host               string
	Port               int
	Password           string
	PoolMinConnections int
	PoolMaxConnections int
	PoolPingThreshold  time.Duration
	Channel            Channel
}

func (o controllerOptions) With(optionSetters ...OptionSetter) controllerOptions {
	for _, os := range optionSetters {
		os(&o)
	}

	return o
}

func defaultOptions(
	host string,
	port int,
	password string,
	channel Channel,
) controllerOptions {
	return controllerOptions{
		Host:     host,
		Port:     port,
		Password: password,

		PoolMinConnections: 1,
		PoolMaxConnections: 16,
		PoolPingThreshold:  time.Minute,

		Channel: channel,
	}
}

// OptionSetter defines an option setter.
type OptionSetter func(*controllerOptions)

// OptionPoolMaxConnections sets maximum idle connections in the pool.
// By default is 16.
func OptionPoolMaxIdleConnections(val int) OptionSetter {
	return func(o *controllerOptions) {
		o.PoolMaxConnections = val
	}
}

// OptionPoolMinIdleConnections sets minimum idle connections in the pool.
// By default is 1.
func OptionPoolMinIdleConnections(val int) OptionSetter {
	return func(o *controllerOptions) {
		o.PoolMinConnections = val
	}
}

// OptionPoolPingThreshold sets minumun ping interval to ensure that
// connection is healthy before getting from the pool.
// By default is 1m. For disabling set 0.
func OptionPoolPingThreshold(val time.Duration) OptionSetter {
	return func(o *controllerOptions) {
		o.PoolPingThreshold = val
	}
}
