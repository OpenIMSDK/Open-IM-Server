package getcdv3

import (
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type RegEtcd struct {
	cli    *clientv3.Client
	ctx    context.Context
	cancel context.CancelFunc
	key    string
}

var rEtcd *RegEtcd

// "%s:///%s/"
func GetPrefix(schema, serviceName string) string {
	return fmt.Sprintf("%s:///%s/", schema, serviceName)
}

// "%s:///%s"
func GetPrefix4Unique(schema, serviceName string) string {
	return fmt.Sprintf("%s:///%s", schema, serviceName)
}

// "%s:///%s/" ->  "%s:///%s:ip:port"
func RegisterEtcd4Unique(schema, etcdAddr, myHost string, myPort int, serviceName string, ttl int) error {
	serviceName = serviceName + ":" + net.JoinHostPort(myHost, strconv.Itoa(myPort))
	return RegisterEtcd(schema, etcdAddr, myHost, myPort, serviceName, ttl)
}

//etcdAddr separated by commas
func RegisterEtcd(schema, etcdAddr, myHost string, myPort int, serviceName string, ttl int) error {
	operationID := utils.OperationIDGenerator()
	args := schema + " " + etcdAddr + " " + myHost + " " + serviceName + " " + utils.Int32ToString(int32(myPort))
	ttl = ttl * 3
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: strings.Split(etcdAddr, ","), DialTimeout: 5 * time.Second})

	log.Info(operationID, "RegisterEtcd args: ", args, ttl)
	if err != nil {
		log.Error(operationID, "clientv3.New failed ", args, ttl, err.Error())
		return fmt.Errorf("create etcd clientv3 client failed, errmsg:%v, etcd addr:%s", err, etcdAddr)
	}

	//  schema:///serviceName/ip:port ->ip:port
	serviceValue := net.JoinHostPort(myHost, strconv.Itoa(myPort))
	serviceKey := GetPrefix(schema, serviceName) + serviceValue

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
	NEWLEASE:
		resp, err := cli.Grant(ctx, int64(ttl))
		if err != nil {
			log.Error(operationID, "Grant failed ", err.Error(), ctx, ttl)
		} else {
			if _, err := cli.Put(ctx, serviceKey, serviceValue, clientv3.WithLease(resp.ID)); err != nil {
				log.Error(operationID, "cli.Put failed ", err.Error(), ctx, args, resp.ID)
			} else {
				kresp, err := cli.KeepAlive(ctx, resp.ID)
				if err != nil {
					log.Error(operationID, "KeepAlive failed ", err.Error(), args, resp.ID)
				} else {
					log.Info(operationID, "RegisterEtcd ok ", args)
					for resp := range kresp {
						log.Debug(operationID, "KeepAlive kresp ok", resp, args)
					}
				}
			}
		}

		log.Error(operationID, "KeepAlive kresp failed ", args)
		time.Sleep(time.Duration(ttl/2) * time.Second)
		goto NEWLEASE
	}()

	rEtcd = &RegEtcd{ctx: ctx,
		cli:    cli,
		cancel: cancel,
		key:    serviceKey}

	return nil
}

func UnRegisterEtcd() {
	//delete
	rEtcd.cancel()
	rEtcd.cli.Delete(rEtcd.ctx, rEtcd.key)
}
