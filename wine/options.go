package wine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Option func(*Options)

type Options struct {
	Name          string
	WineBin       string
	WineTricksBin string
	DefaultPrefix string
}

func ApplyOptions(opts ...Option) Options {
	o := Options{
		Name:          "wine",
		WineBin:       "wine",
		WineTricksBin: "winetricks",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return o
}

func NewFromOptions(o Options) *Runner {
	return &Runner{Options: o}
}

func Load(path string) (*Runner, error) {
	o, err := LoadOptions(path)
	if err != nil {
		return nil, err
	}

	return NewFromOptions(o), nil
}

func LoadOptions(path string) (Options, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Options{}, fmt.Errorf("read wine options: %w", err)
	}

	var o Options
	if err := json.Unmarshal(data, &o); err != nil {
		return Options{}, fmt.Errorf("decode wine options: %w", err)
	}

	return o, nil
}

func Delete(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete wine options: %w", err)
	}

	return nil
}

func (o Options) Save(path string) error {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return fmt.Errorf("encode wine options: %w", err)
	}

	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create wine options directory: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write wine options: %w", err)
	}

	return nil
}

func WithName(name string) Option {
	return func(r *Options) {
		if name != "" {
			r.Name = name
		}
	}
}

func WithWineBin(path string) Option {
	return func(r *Options) {
		if path != "" {
			r.WineBin = path
		}
	}
}

func WithWineTricksBin(path string) Option {
	return func(r *Options) {
		if path != "" {
			r.WineTricksBin = path
		}
	}
}

func WithDefaultPrefix(prefix string) Option {
	return func(r *Options) {
		r.DefaultPrefix = prefix
	}
}
