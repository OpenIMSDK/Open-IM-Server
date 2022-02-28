package msg

import pbChat "Open_IM/pkg/proto/chat"

// BeforeSendFilter handles custom logic before send msg.
type BeforeSendFilter func(ctx *SendContext, pb *pbChat.SendMsgReq) (*pbChat.SendMsgResp, bool, error)

// AfterSendFilter handles custom logic after send msg.
type AfterSendFilter func(ctx *SendContext, req *pbChat.SendMsgReq, res *pbChat.SendMsgResp) (*pbChat.SendMsgResp, bool, error)
