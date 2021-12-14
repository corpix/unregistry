package revip

import (
	"net/url"

	etcd "go.etcd.io/etcd/clientv3"
	pb "go.etcd.io/etcd/mvcc/mvccpb"
)

const (
	EtcdPathDelimiter   = "/"
	etcdOperationPut    = int32(pb.PUT)
	etcdOperationDelete = int32(pb.DELETE)
)

type etcdUpdateEvent struct {
	operation int32
	data      []byte
	key       string
	version   int64
}

// NewEtcdClient creates etcd client from URL string.
func NewEtcdClient(u string) (*etcd.Client, error) {
	uu, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	return NewEtcdClientFromURL(uu)
}

// NewEtcdClient creates etcd client from *url.URL.
func NewEtcdClientFromURL(uu *url.URL) (*etcd.Client, error) {
	password, _ := uu.User.Password()
	c := etcd.Config{
		Endpoints: []string{uu.Host},
		Username:  uu.User.Username(),
		Password:  password,
	}
	return etcd.New(c)
}
