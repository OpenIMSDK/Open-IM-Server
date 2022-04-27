/*
** description("").
** copyright('open-im,www.open-im.io').
** author("fg,Gordon@open-im.io").
** time(2021/3/22 15:33).
 */
package logic

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/kafka"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	"Open_IM/pkg/statistics"
	"fmt"
)

var (
	watcher      *getcdv3.Watcher
	rpcServer    RPCServer
	pushCh       PushConsumerHandler
	pushTerminal []int32
	producer     *kafka.Producer
	count        uint64
)

func Init(rpcPort int) {
	watcher = getcdv3.NewWatcher()
	rpcServer.Init(rpcPort)
	pushCh.Init()
	pushTerminal = []int32{constant.IOSPlatformID, constant.AndroidPlatformID}
}
func init() {
	producer = kafka.NewKafkaProducer(config.Config.Kafka.Ws2mschat.Addr, config.Config.Kafka.Ws2mschat.Topic)
	statistics.NewStatistics(&count, config.Config.ModuleName.PushName, fmt.Sprintf("%d second push to msg_gateway count", 300), 300)
}

func Run() {
	go rpcServer.run()
	go pushCh.pushConsumerGroup.RegisterHandleAndConsumer(&pushCh)
}
