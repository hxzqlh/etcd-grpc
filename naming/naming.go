package naming

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
)

func Register(etcdAddr, name, addr, schema string, ttl int64) error {
	var err error
	if cli == nil {
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(etcdAddr, ";"),
			DialTimeout: 15 * time.Second,
		})
		if err != nil {
			log.Printf("connect to etcd err:%s", err)
			return err
		}
	}

	ticker := time.NewTicker(time.Second * time.Duration(ttl))

	go func() {
		for {
			getResp, err := cli.Get(context.Background(), "/"+schema+"/"+name+"/"+addr)
			if err != nil {
				log.Printf("getResp:%+v\n", getResp)
				log.Printf("Register:%s", err)
			} else if getResp.Count == 0 {
				err = withAlive(name, addr, schema, ttl)
				if err != nil {
					log.Printf("keep alive:%s", err)
				}
			}
			<-ticker.C
		}
	}()
	return nil
}

// withAlive 创建租约
func withAlive(name, addr, schema string, ttl int64) error {
	leaseResp, err := cli.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}

	log.Printf("key:%v\n", "/"+schema+"/"+name+"/"+addr)
	_, err = cli.Put(context.Background(), "/"+schema+"/"+name+"/"+addr, addr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		log.Printf("put etcd error:%s", err)
		return err
	}

	ch, err := cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		log.Printf("keep alive error:%s", err)
		return err
	}

	// 清空 keep alive 返回的channel
	go func() {
		for {
			<-ch
		}
	}()

	return nil
}

func UnRegister(name, addr, schema string) {
	if cli != nil {
		cli.Delete(context.Background(), "/"+schema+"/"+name+"/"+addr)
	}
}
