package configmodel

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"
)

// Static is read once at startup; changing it requires a restart.
type Static struct {
	Log         Log                   `yaml:"log"`
	EntryPoints map[string]EntryPoint `yaml:"entryPoints"`
	API         API                   `yaml:"api"`
	Providers   Providers             `yaml:"providers"`
	LifeCycle   LifeCycle             `yaml:"lifeCycle"`
}

type Log struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type EntryPoint struct {
	Address  string   `yaml:"address"`
	Timeouts Timeouts `yaml:"timeouts"`
}

// Timeouts guard the listener against slow clients (Slowloris and friends).
// Write defaults to 0 so long-lived responses (streaming, SSE) are not cut off.
type Timeouts struct {
	ReadHeader time.Duration `yaml:"readHeader"`
	Read       time.Duration `yaml:"read"`
	Write      time.Duration `yaml:"write"`
	Idle       time.Duration `yaml:"idle"`
}

// API is the read-only admin endpoint; an empty Address disables it.
type API struct {
	Address string `yaml:"address"`
}

type Providers struct {
	File *FileProvider `yaml:"file"`
}

type FileProvider struct {
	Path string `yaml:"path"`
}

type LifeCycle struct {
	GraceTimeout time.Duration `yaml:"graceTimeout"`
}

func DefaultTimeouts() Timeouts {
	return Timeouts{
		ReadHeader: 10 * time.Second,
		Read:       60 * time.Second,
		Idle:       180 * time.Second,
	}
}

func (s *Static) SetDefaults() {
	if s.Log.Level == "" {
		s.Log.Level = "info"
	}
	if s.Log.Format == "" {
		s.Log.Format = "json"
	}
	if s.LifeCycle.GraceTimeout == 0 {
		s.LifeCycle.GraceTimeout = 10 * time.Second
	}

	def := DefaultTimeouts()
	for name, ep := range s.EntryPoints {
		if ep.Timeouts.ReadHeader == 0 {
			ep.Timeouts.ReadHeader = def.ReadHeader
		}
		if ep.Timeouts.Read == 0 {
			ep.Timeouts.Read = def.Read
		}
		if ep.Timeouts.Idle == 0 {
			ep.Timeouts.Idle = def.Idle
		}
		s.EntryPoints[name] = ep
	}
}

func (s *Static) Validate() error {
	var errs []error

	if !slices.Contains([]string{"debug", "info", "warn", "error"}, s.Log.Level) {
		errs = append(errs, fmt.Errorf("log.level %q: must be debug, info, warn or error", s.Log.Level))
	}
	if !slices.Contains([]string{"json", "text"}, s.Log.Format) {
		errs = append(errs, fmt.Errorf("log.format %q: must be json or text", s.Log.Format))
	}

	if len(s.EntryPoints) == 0 {
		errs = append(errs, errors.New("entryPoints: at least one entry point is required"))
	}
	for _, name := range slices.Sorted(maps.Keys(s.EntryPoints)) {
		if s.EntryPoints[name].Address == "" {
			errs = append(errs, fmt.Errorf("entryPoints.%s.address: required", name))
		}
	}

	if f := s.Providers.File; f != nil && f.Path == "" {
		errs = append(errs, errors.New("providers.file.path: required"))
	}

	return errors.Join(errs...)
}
