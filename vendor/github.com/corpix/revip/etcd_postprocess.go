package revip

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	etcd "go.etcd.io/etcd/clientv3"
)

//

// WithUpdatesFromEtcdConfig represents a configuration for WithUpdatesFromEtcd.
// By default OnError panics.
type UpdateFromEtcdConfig struct {
	Ctx           context.Context
	BatchSize     int
	BatchDuration time.Duration
	OnError       func(error)
}

// UpdateFromEtcdOption represents an option for WithUpdatesFromEtcdConfig.
type UpdateFromEtcdOption = func(*UpdateFromEtcdConfig)

// UpdatesFromEtcdContext set Ctx on UpdateFromEtcdConfig.
func UpdatesFromEtcdContext(ctx context.Context) UpdateFromEtcdOption {
	return func(c *UpdateFromEtcdConfig) { c.Ctx = ctx }
}

// UpdatesFromEtcdBatch set Batch* on UpdateFromEtcdConfig.
func UpdatesFromEtcdBatch(size int, duration time.Duration) UpdateFromEtcdOption {
	return func(c *UpdateFromEtcdConfig) {
		c.BatchSize = size
		c.BatchDuration = duration
	}
}

// UpdatesFromEtcdErrorHandler set OnError on UpdateFromEtcdConfig.
func UpdatesFromEtcdErrorHandler(cb func(error)) UpdateFromEtcdOption {
	return func(c *UpdateFromEtcdConfig) {
		c.OnError = cb
	}
}

//

// etcdWatchPump is a background job for UpdateFromEtcdConfig.
// It pumps events from etcd client watcher to the single events channel,
// which is handled by etcdWatchHandle.
func etcdWatchPump(ctx context.Context, ch etcd.WatchChan, events chan etcdUpdateEvent) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case v, ok := <-ch:
			if !ok {
				break loop
			}

			for _, evt := range v.Events {
				events <- etcdUpdateEvent{
					operation: int32(evt.Type),
					data:      evt.Kv.Value,
					key:       string(evt.Kv.Key),
					version:   evt.Kv.Version,
				}
			}
		}
	}
}

// etcdWatchHandle is a background job for UpdateFromEtcdConfig.
// Implements batching and notifications for etcdWatchPump events (via calling .Update(...) on Config struct).
func etcdWatchHandle(ctx context.Context, batchSize int, batchDuration time.Duration, c Config, namespace string, events chan etcdUpdateEvent, v Updateable, onError func(error), f Unmarshaler) {
	var (
		versions = map[string]int64{}
		acc      = make([]etcdUpdateEvent, batchSize)
		n        = 0
		err      error
	)

	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-events:
			if !ok {
				return
			}
			if ver, ok := versions[evt.key]; ok && ver >= evt.version {
				continue // skip already updated keys
			}

			acc[n] = evt
			n++

			if n >= batchSize {
				goto flush
			}
		case <-time.After(batchDuration):
			goto flush
		}
		continue

	flush:
		if n == 0 {
			continue
		}

		dst := reflect.New(reflect.ValueOf(c).Elem().Type()).Interface()
		err = mapstructure.Decode(c, &dst)
		if err != nil {
			onError(err)
			continue
		}

		evtByKey := map[string]etcdUpdateEvent{}
		for _, evt := range acc[:n] {
			evtByKey[evt.key] = evt
		}

		err = walkStruct(dst, func(v reflect.Value, path []string) error {
			k := v.Type().Kind()
			switch { // ignore nil's and substructs (substructs may have their own handlers)
			case k == reflect.Struct:
				return skipBranch
			case k == reflect.Ptr:
				return nil
			case !v.CanAddr():
				return skipBranch
			}

			key := strings.Join(prefixPath(namespace, path), EtcdPathDelimiter)
			if evt, ok := evtByKey[key]; ok {
				switch evt.operation {
				case etcdOperationDelete:
					v.Set(reflect.New(v.Type()).Elem())
				case etcdOperationPut:
					switch indirectType(v.Type()).Kind() {
					case reflect.Map: // erase map because unmarshal update semantics is "merge"
						v.Set(reflect.New(v.Type()).Elem())
					}

					err := f(evt.data, v.Addr().Interface())
					if err != nil {
						return err
					}
				}
			}

			return nil
		})
		if err != nil {
			onError(err)
			continue
		}

		// update configuration

		err = v.Update(dst)
		if err != nil {
			onError(err)
			continue
		}

		err = mapstructure.Decode(dst, c)
		if err != nil {
			onError(err)
			continue
		}

		//

		n = 0
	}
}

// WithUpdatesFromEtcd represents a postprocess Option which handles updates from etcd.
func WithUpdatesFromEtcd(client *etcd.Client, namespace string, f Unmarshaler, op ...UpdateFromEtcdOption) Option {
	cfg := &UpdateFromEtcdConfig{
		Ctx:           context.Background(),
		BatchSize:     64,
		BatchDuration: 1 * time.Second,
		OnError:       func(err error) { panic(err) },
	}
	for _, apply := range op {
		apply(cfg)
	}

	return func(c Config, m ...OptionMeta) error {
		v, ok := c.(Updateable)
		if !ok {
			return nil
		}

		events := make(chan etcdUpdateEvent, 16)

		go etcdWatchPump(cfg.Ctx, client.Watch(cfg.Ctx, namespace, etcd.WithPrefix()), events)
		go etcdWatchHandle(cfg.Ctx, cfg.BatchSize, cfg.BatchDuration, c, namespace, events, v, cfg.OnError, f)

		return nil
	}
}
