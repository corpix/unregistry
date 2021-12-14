# revip

Dead-simple configuration loader.

It supports:

- JSON, TOML, YAML and you could add your own format unmarshaler (see `Unmarshaler` type)
- file, reader and environment sources support, also you could add your own (see `Option` type and `sources.go`)
- extendable postprocessing support (defaults, validation, expansion, see `Option` type and `postprocess.go`)
- dot-notation to access configuration keys
- reading and writing to and from etcd (with watchers)

[Godoc](https://godoc.org/github.com/corpix/revip)

---

### example


### run

```console
$ go run ./example/basic/main.go
(main.Config) {
 Foo: (*main.Foo)(0xc00000e540)({
  Bar: (string) (len=3) "bar",
  Qux: (bool) false
 }),
 Baz: (int) 666,
 Dox: ([]string) <nil>,
 Box: ([]int) (len=3 cap=3) {
  (int) 666,
  (int) 777,
  (int) 888
 },
 Fox: (map[string]*main.Foo) (len=1) {
  (string) (len=3) "key": (*main.Foo)(0xc00000e960)({
   Bar: (string) (len=13) "default value",
   Qux: (bool) false
  })
 },
 Gox: ([]*main.Foo) (len=1 cap=1) {
  (*main.Foo)(0xc00000e980)({
   Bar: (string) (len=13) "default value",
   Qux: (bool) false
  })
 },
 key: (string) (len=25) "value written by Expand()"
}
```

Other things to try:

```console
$ REVIP_FOO_BAR=hello go run ./example/basic/main.go
$ REVIP_BOX=888,777,666 go run ./example/basic/main.go
$ REVIP_BAZ=0 go run ./example/basic/main.go
```

## license

[public domain](https://unlicense.org/)
