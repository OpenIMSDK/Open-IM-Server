package filters

import (
	rpcChat "Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/constant"
	pbChat "Open_IM/pkg/proto/chat"
	"context"
	"errors"
	"fmt"
	"time"
)

func MockBeforeSendFilter1(ctx *rpcChat.SendContext, pb *pbChat.SendMsgReq) (*pbChat.SendMsgResp, bool, error) {

	cp := ctx.Copy()

	// 假设这里我们发起一次rpc，超时时间 1s
	sc, cancel := context.WithTimeout(cp, time.Second*1)
	defer cancel()

	sc.Value("ss") // 无意义，只是为了不报错

	// callRpc(sc,pb)

	// other logic
	// doOther()

	// 模拟处理时间
	time.Sleep(2 * time.Second)

	// mock 将一些需要的数据放入ctx，供后续其他拦截过滤器使用
	ctxKey := "test_key"
	v := true
	// fmt.Printf("MockBeforeSendFilter1:%s set value to ctx,value is :%v\n", ctxKey, v)
	ctx.WithValue(ctxKey, v)

	return nil, true, nil
}

// MockBeforeSendFilter is a mock handle that handles custom logic before send msg.
func MockBeforeSendFilter2(ctx *rpcChat.SendContext, pb *pbChat.SendMsgReq) (*pbChat.SendMsgResp, bool, error) {
	// 取出放进去的测试数据
	ctxKey := "test_key"
	v, ok := ctx.Value(ctxKey).(bool)
	if ok {
		fmt.Printf("MockBeforeSendFilter2:%s selected from ctx,value is :%v\n", ctxKey, v)
	}

	fmt.Printf("MockBeforeSendFilter2 trigger,contentType:%d\n", pb.MsgData.GetContentType())
	if pb.MsgData.ContentType == constant.Text {
		msg := string(pb.MsgData.Content)
		// fmt.Printf("text msg:%s\n", msg)
		if msg == "this is a m..m..mock msg" {
			fmt.Println(".==>msg had banned")
			return nil, false, errors.New("BANG! This msg has been banned by MockBeforeSendHandler")
		}
	}

	return nil, true, nil
}

// MockAfterSendFilter is a mock handle that handles custom logic after send msg.
func MockAfterSendFilter(ctx *rpcChat.SendContext, pb *pbChat.SendMsgReq, res *pbChat.SendMsgResp) (*pbChat.SendMsgResp, bool, error) {
	ctxKey := "test_key"
	v, ok := ctx.Value(ctxKey).(bool)
	if ok {
		fmt.Printf("MockAfterSendFilter:%s selected from ctx,value is :%v\n", ctxKey, v)
	}

	fmt.Printf("MockAfterSendFilter trigger,contentType:%d\n", pb.MsgData.GetContentType())

	return res, true, nil
}
