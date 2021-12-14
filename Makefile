.DEFAULT_GOAL := all

## parameters

name                 = unregistry
group                = corpix
remote               = git.backbone
namespace            = $(remote)/$(group)
version             ?= development
os                  ?=
binary              ?= ./main
args                ?=
container_namespace ?= ghcr.io/$(group)/$(name)
container_tag       ?= latest
docker_user         ?=
docker_password     ?=

## bindings

root                := $(patsubst %/,%,$(dir $(realpath $(firstword $(MAKEFILE_LIST)))))
nix_dir             := nix
pkg_prefix          := $(namespace)/$(name)
tmux                := tmux -2 -f $(root)/.tmux.conf -S $(root)/.tmux
tmux_session        := $(name)
nix                 := nix --show-trace $(nix_opts)

### reusable and long opts for commands inside rules

shell_opts = -v nix:/nix:rw                     \
	-v $(root):/chroot                      \
	-e COLUMNS=$(COLUMNS)                   \
	-e LINES=$(LINES)                       \
	-e TERM=$(TERM)                         \
	-e NIX_BUILD_CORES=$(NIX_BUILD_CORES)   \
	-e HOME=/chroot                         \
	-w /chroot                              \
	--hostname localhost                    \
	$(foreach v,$(ports), -p $(v):$(v) )

## helpers

, = ,

## macro

define fail
{ echo "error: "$(1) 1>&2; exit 1; }
endef

## targets

.PHONY: all
all: build # test, check and build all cmds

.PHONY: help
help: # print defined targets and their comments
	@grep -Po '^[a-zA-Z%_/\-\s]+:+(\s.*$$|$$)' Makefile \
		| sort                                      \
		| sed 's|:.*#|#|;s|#\s*|#|'                 \
		| column -t -s '#' -o ' | '

### releases

.PHONY: nix/build/container
nix/build/container build/container.tar.gz: # build container with nix (in userspace)
	$(nix) build -o build/container.tar.gz     \
        --argstr name       $(name)                \
        --argstr namespace  $(container_namespace) \
        --argstr version    $(version)             \
        --argstr tag        $(container_tag)       \
        -f $(nix_dir)/container.nix

.PHONY: nix/push/container
nix/push/container: build/container.tar.gz # upload container built by nix
	@# about insecure policy, see: https://github.com/containers/skopeo/issues/394
	@skopeo --insecure-policy                                   \
		copy --dest-creds=$(docker_user):$(docker_password) \
		docker-archive://$(root)/$<                         \
		docker://$(container_namespace)/$(name):$(container_tag)

### development

.PHONY: build
build $(binary): # build application `binary`
	swag init --generalInfo server.go --dir pkg/telemetry/ --output pkg/telemetry/
	GOOS=$(os)                                                       \
	go build -o $(binary)                                            \
		--ldflags "                                              \
			-X $(pkg_prefix)/pkg/meta.Version=$(version)     \
		"                                                        \
		./main.go

.PHONY: fmt
fmt: # run go fmt
	go fmt ./...

.PHONY: tidy
tidy: # run go mod tidy
	go mod tidy

.PHONY: vendor
vendor: # run go mod vendor
	go mod vendor

.PHONY: run
run: build # run application
	$(binary)

#### testing

.PHONY: test
test: # run unit tests
	go test -v ./...

.PHONY: lint
lint: # run linter
	golangci-lint --color=always --timeout=120s run ./...

.PHONY: profile
profile: build # collect profile for application
	$(binary) --profile --duration=30s

.PHONY: trace
trace: build # collect trace for application
	$(binary) --trace --duration=30s

.PHONY: pprof
pprof: $(binary) # run pprof web server to visualize collected `profile`
	@[ -z "$(profile)" ] && $(call fail,"profile=<value> parameter is required$(,) available values: cpu$(,) heap") || true
	go tool pprof -http=":8081" $(binary) $(profile).prof

#### runners

## env

.PHONY: run/shell
run/shell: # enter development environment with nix-shell
	nix-shell

.PHONY: run/cage/shell
run/cage/shell: # enter sandboxed development environment with nix-cage
	nix-cage

.PHONY: run/nix/repl
run/nix/repl: # run nix repl for nixpkgs from env
	nix repl '<nixpkgs>'

## dev session

.PHONY: run/tmux/session
run/tmux/session: # start development environment
	@$(tmux) has-session -t $(tmux_session) && $(call fail,tmux session $(tmux_session) already exists$(,) use: '$(tmux) attach-session -t $(tmux_session)' to attach) || true
	@$(tmux) new-session -s $(tmux_session) -n console -d
	@while ! $(tmux) select-window -t $(tmux_session):0; do sleep 0.5; done

	@if [ -f $(root)/.personal.tmux.conf ]; then             \
		$(tmux) source-file $(root)/.personal.tmux.conf; \
	fi

	@$(tmux) attach-session -t $(tmux_session)

.PHONY: run/tmux/attach
run/tmux/attach: # attach to development session if running
	@$(tmux) attach-session -t $(tmux_session)

.PHONY: run/tmux/kill
run/tmux/kill: # kill development environment
	@$(tmux) kill-session -t $(tmux_session)

#### runners

.PHONY: run/docker/shell
run/docker/shell: # run development environment shell
	@docker run --rm -it                   \
		--log-driver=none              \
		$(shell_opts) nixos/nix:latest \
		nix-shell --run 'exec make run/shell'

.PHONY: run/docker/clean
run/docker/clean: # clean development environment artifacts
	docker volume rm nix

test/prometheus/data: # make sure prometheus data directory exists
	mkdir -p $@

.PHONY: run/prometheus
run/prometheus: test/prometheus/data # run prometheus metrics collection service
	@bash -xec "cd $(dir $<); exec prometheus --config.file=./prometheus.yml --storage.tsdb.path=./data"

clean:: # clean prometheus state
	rm -rf test/prometheus/data

##

.PHONY: clean
clean:: # clean state
	rm -rf result*
	rm -rf build main
	rm -rf .cache/* .local/* .config/* || true
