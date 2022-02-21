package memory

import "github.com/milosgajdos/orbnet/pkg/graph/style"

// Options configure graph.
type Options struct {
	// UID configures UID
	UID string
	// Label configures Label.
	Label string
	// Attrs configures Attrs.
	Attrs map[string]interface{}
	// DotID configures DOT ID
	DotID string
	// Type is graph type
	Type string
	// Weight configures weight.
	Weight float64
	// Style configures style.
	Style style.Style
}

// Option is functional graph option.
type Option func(*Options)

// WithUID sets UID option.
func WithUID(u string) Option {
	return func(o *Options) {
		o.UID = u
	}
}

// WithLabel sets Label option.
func WithLabel(l string) Option {
	return func(o *Options) {
		o.Label = l
	}
}

// WithAttrs sets Attrs option,
func WithAttrs(a map[string]interface{}) Option {
	return func(o *Options) {
		o.Attrs = a
	}
}

// WithDotID sets DotID option.
func WithDotID(d string) Option {
	return func(o *Options) {
		o.DotID = d
	}
}

// WithType sets Type option.
func WithType(t string) Option {
	return func(o *Options) {
		o.Type = t
	}
}

// WithWeight sets Weight option.
func WithWeight(w float64) Option {
	return func(o *Options) {
		o.Weight = w
	}
}

// WithStyle sets Style option.
func WithStyle(s style.Style) Option {
	return func(o *Options) {
		o.Style = s
	}
}
