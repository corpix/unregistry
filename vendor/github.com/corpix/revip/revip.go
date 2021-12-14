package revip

import (
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

const PathDelimiter = "."

const (
	SchemeEmpty   = ""
	SchemeFile    = "file" // file://./config.yml
	SchemeEnviron = "env"  // env://prefix
	SchemeEtcd    = "etcd" // etcd://user@password:127.0.0.1:2379/namespace
)

var (
	// FromSchemes represents schemes supported for sources.
	FromSchemes = []string{
		SchemeFile,
		SchemeEnviron,
		SchemeEtcd,
	}
	// ToSchemes represents schemes supported for destrinations.
	ToSchemes = []string{
		SchemeFile,
		SchemeEtcd,
	}
)

// Config is a configuration represented by user-specified type.
type Config = interface{}

// Option defines generic interface for configuration source.
type Option = func(c Config, m ...OptionMeta) error

// OptionMeta is an optional meta-data to be passed to `Option`.
type OptionMeta interface{}

// Defaultable is an interface which any `Config` could implement
// to define a custom default values for sub-tree it owns.
type Defaultable interface {
	Default()
}

// Validatable is an interface which any `Config` could implement
// to define a validation rules for sub-tree it owns.
type Validatable interface {
	Validate() error
}

// Expandable is an interface which any `Config` could implement
// to define an expansion rules for sub-tree it owns.
type Expandable interface {
	Expand() error
}

// Updateable is an interface which any `Config` could implement
// to define a custom configuration update logic.
type Updateable interface {
	Update(Config) error
}

// Revip represents configuration loaded by `Load`.
type Revip struct {
	// config represents configuration data, it should always be a pointer.
	config Config
}

// Unwrap returns a pointer to the inner configuration data structure.
func (r *Revip) Unwrap() interface{} { return r.config }

// Copy writes a shallow copy of the configuration into `v`.
func (r *Revip) Copy(v interface{}) error {
	return mapstructure.WeakDecode(r.config, v)
}

// DeepCopy writes a deep copy of the configuration into `v`.
func (r *Revip) DeepCopy(v interface{}) error {
	return mapstructure.Decode(r.config, v)
}

// Path uses dot notation to retrieve substruct addressable by `path` or
// return an error if key was not found(`ErrNotFound`) or
// something gone terribly wrong.
func (r *Revip) Path(dst Config, path string) error {
	found := false

	err := walkStruct(r.config, func(v reflect.Value, xs []string) error {
		if strings.Join(xs, PathDelimiter) == path {
			found = true
			err := mapstructure.WeakDecode(v.Interface(), dst)
			if err != nil {
				return err
			}
			return stopIteration
		}

		return nil
	})
	if err != nil {
		return err
	}

	if !found {
		return &ErrPathNotFound{Path: path}
	}

	return nil
}

// New wraps configuration represented by `c` with come useful methods.
func New(c Config) *Revip {
	if reflect.TypeOf(c).Kind() != reflect.Ptr {
		panic("config must be a pointer")
	}

	return &Revip{config: c}
}

// LoadWithMeta feeds into each `op` `meta` and applies `op` in order to
// fill the configuration in `v` then constructs a `*Revip` data-structure.
func LoadWithMeta(v Config, meta []OptionMeta, op ...Option) (*Revip, error) {
	var err error
	for _, f := range op {
		err = f(v)
		if err != nil {
			return nil, err
		}
	}

	return New(v), nil
}

// Load applies each `op` in order to fill the configuration in `v` and
// constructs a `*Revip` data-structure.
func Load(v Config, op ...Option) (*Revip, error) {
	return LoadWithMeta(v, nil, op...)
}
