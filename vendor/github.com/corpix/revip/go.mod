module github.com/corpix/revip

go 1.15

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pelletier/go-toml v1.6.0
	github.com/stretchr/testify v1.4.0
	go.etcd.io/etcd v0.5.0-alpha.5.0.20210226220824-aa7126864d82
)

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
