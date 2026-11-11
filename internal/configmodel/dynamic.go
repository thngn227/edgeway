package configmodel

import (
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
)

// Dynamic describes routing and can change at runtime without a restart.
// It is the contract shared by every provider (file today, control plane later).
type Dynamic struct {
	Routers  map[string]Router  `yaml:"routers"`
	Services map[string]Service `yaml:"services"`
}

type Router struct {
	// EntryPoints empty means the router listens on every entry point.
	EntryPoints []string `yaml:"entryPoints"`
	Rule        string   `yaml:"rule"`
	Priority    int      `yaml:"priority"`
	Service     string   `yaml:"service"`
}

type Service struct {
	LoadBalancer *LoadBalancer `yaml:"loadBalancer"`
}

type LoadBalancer struct {
	Servers []Server `yaml:"servers"`
}

type Server struct {
	URL string `yaml:"url"`
}

// Validate checks internal references and that routers only use known entry points.
func (d *Dynamic) Validate(entryPoints map[string]EntryPoint) error {
	var errs []error

	for _, name := range slices.Sorted(maps.Keys(d.Routers)) {
		r := d.Routers[name]
		if r.Rule == "" {
			errs = append(errs, fmt.Errorf("routers.%s.rule: required", name))
		}
		if _, ok := d.Services[r.Service]; !ok {
			errs = append(errs, fmt.Errorf("routers.%s.service: unknown service %q", name, r.Service))
		}
		for _, ep := range r.EntryPoints {
			if _, ok := entryPoints[ep]; !ok {
				errs = append(errs, fmt.Errorf("routers.%s.entryPoints: unknown entry point %q", name, ep))
			}
		}
	}

	for _, name := range slices.Sorted(maps.Keys(d.Services)) {
		lb := d.Services[name].LoadBalancer
		if lb == nil || len(lb.Servers) == 0 {
			errs = append(errs, fmt.Errorf("services.%s.loadBalancer.servers: at least one server is required", name))
			continue
		}
		for i, srv := range lb.Servers {
			if err := validateServerURL(srv.URL); err != nil {
				errs = append(errs, fmt.Errorf("services.%s.loadBalancer.servers[%d].url: %w", name, i, err))
			}
		}
	}

	return errors.Join(errs...)
}

func validateServerURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%q: scheme must be http or https", raw)
	}
	if u.Host == "" {
		return fmt.Errorf("%q: host is required", raw)
	}
	return nil
}
