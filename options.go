package logze

import "github.com/rs/zerolog"

type Option func(*Config)

func WithSampler(sampler zerolog.Sampler) Option {
	return func(c *Config) {
		c.Sampler = sampler
	}
}
