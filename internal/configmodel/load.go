package configmodel

import (
	"errors"
	"fmt"
	"io"
	"os"

	"go.yaml.in/yaml/v3"
)

func LoadStatic(path string) (*Static, error) {
	var s Static
	if err := decodeFile(path, &s); err != nil {
		return nil, err
	}
	s.SetDefaults()
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid static config %s:\n%w", path, err)
	}
	return &s, nil
}

func LoadDynamic(path string, entryPoints map[string]EntryPoint) (*Dynamic, error) {
	var d Dynamic
	if err := decodeFile(path, &d); err != nil {
		return nil, err
	}
	if err := d.Validate(entryPoints); err != nil {
		return nil, fmt.Errorf("invalid dynamic config %s:\n%w", path, err)
	}
	return &d, nil
}

func decodeFile(path string, out any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	// A typo in a key must fail loudly instead of being silently ignored.
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	return nil
}
