package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/resolver"
)

type etcdBuilder struct {
}

const (
	scheme = "dns"
)

type etcdResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
	client *clientv3.Client
}

var etcdCli *clientv3.Client

func EtcdRegister(cli *clientv3.Client) {
	etcdCli = cli
	resolver.Register(&etcdBuilder{})
}

func (e etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	er := &etcdResolver{
		target: target,
		cc:     cc,
		client: etcdCli,
	}
	go er.watch()
	return er, nil
}

func (e etcdResolver) watch() {
	keyPrefix := fmt.Sprintf("%s://%s", e.target.Scheme, e.target.Authority)
	fmt.Println(keyPrefix)
	resp, err := e.client.Get(context.TODO(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	addresses := make([]resolver.Address, 0)
	var addr resolver.Address
	for _, kv := range resp.Kvs {
		err := json.Unmarshal(kv.Value, &addr)
		if err != nil {
			panic(err)
		}
		addresses = append(addresses, addr)
	}
	var state resolver.State
	state = resolver.State{
		Addresses:     addresses,
		ServiceConfig: nil,
	}
	e.cc.UpdateState(state)

	wc := e.client.Watch(context.TODO(), keyPrefix, clientv3.WithPrefix())
	for response := range wc {
		for _, event := range response.Events {
			_ = json.Unmarshal(event.Kv.Value, &addr)
			fmt.Printf("收到事件：%#v\n", event)
			switch event.Type {
			case mvccpb.PUT:
				addresses = append(addresses, addr)
				state = resolver.State{
					Addresses:     addresses,
					ServiceConfig: nil,
				}
				break
			case mvccpb.DELETE:
				for k, r := range addresses {
					if r == addr {
						addresses = append(addresses[:k], addresses[k+1:]...)
					}
				}
				break
			}
			e.cc.UpdateState(state)
		}
	}
}

func (e etcdBuilder) Scheme() string {
	return scheme
}

// Target constructs a endpoint resolver target.
func Target(id, endpoint string) string {
	return fmt.Sprintf("%s://%s/%s", scheme, id, endpoint)
}

func TargetPrefix(id string) string {
	return fmt.Sprintf("%s://%s/", scheme, id)
}

func (e etcdResolver) ResolveNow(option resolver.ResolveNowOption) {

}
func (e etcdResolver) Close() {

}
