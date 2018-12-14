package main

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"context"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"sync"
	"log"
)

type ClientDis struct {
	client        *clientv3.Client
	serverList    map[string]string
	lock          sync.Mutex
}

func NewClientDis (addr []string)( *ClientDis, error){
	conf := clientv3.Config{
		Endpoints:   addr,
		DialTimeout: 5 * time.Second,
	}
	if client, err := clientv3.New(conf); err == nil {
		return &ClientDis{
			client:client,
			serverList:make(map[string]string),
		}, nil
	} else {
		return nil ,err
	}
}


func (this * ClientDis) GetService(prefix string) ([]string ,error){
	resp, err := this.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	addrs := this.extractAddrs(resp)

    go this.watcher(prefix)
	return addrs ,nil
}


func (this *ClientDis) watcher(prefix string) {
	rch := this.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				this.SetServiceList(string(ev.Kv.Key),string(ev.Kv.Value))
			case mvccpb.DELETE:
				this.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
}

func (this *ClientDis) extractAddrs(resp *clientv3.GetResponse) []string {
	addrs := make([]string,0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			this.SetServiceList(string(resp.Kvs[i].Key),string(resp.Kvs[i].Value))
			addrs = append(addrs, string(v))
		}
	}
	return addrs
}

func (this *ClientDis) SetServiceList(key,val string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.serverList[key] = string(val)
	log.Println("set data key :",key,"val:",val)
}

func (this *ClientDis) DelServiceList(key string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.serverList,key)
	log.Println("del data key:", key)
}


func (this *ClientDis) SerList2Array()[]string {
	this.lock.Lock()
	defer this.lock.Unlock()
	addrs := make([]string,0)

	for _, v := range this.serverList {
		addrs = append(addrs,v)
	}
	return addrs
}