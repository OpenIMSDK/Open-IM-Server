package msg

import (
	pbChat "Open_IM/pkg/proto/chat"
	"context"
	"net/http"
	"time"
)

var _ context.Context = (*SendContext)(nil)

// SendContext is the most important part of RPC SendMsg. It allows us to pass variables between middleware
type SendContext struct {
	ctx context.Context
	rpc *rpcChat
	// beforeFilters are filters which will be triggered before send msg
	beforeFilters []BeforeSendFilter
	// afterSenders are filters which will be triggered after send msg
	afterSenders []AfterSendFilter
}

func NewSendContext(ctx context.Context, rpc *rpcChat) *SendContext {
	return &SendContext{
		ctx:           ctx,
		rpc:           rpc,
		beforeFilters: rpc.beforeSenders,
		afterSenders:  rpc.afterSenders,
	}
}

// doBeforeFilters executes the pending filters in the chain inside the calling handler.
func (c *SendContext) doBeforeFilters(pb *pbChat.SendMsgReq) (*pbChat.SendMsgResp, bool, error) {
	for _, handler := range c.beforeFilters {
		res, ok, err := handler(c, pb)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return res, ok, nil
		}
	}

	return nil, true, nil
}

// doAfterFilters executes the pending filters in the chain inside the calling handler.
func (c *SendContext) doAfterFilters(req *pbChat.SendMsgReq, res *pbChat.SendMsgResp) (*pbChat.SendMsgResp, bool, error) {
	for _, handler := range c.afterSenders {
		res, ok, err := handler(c, req, res)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return res, ok, nil
		}
	}

	return nil, true, nil
}

func (c *SendContext) SendMsg(pb *pbChat.SendMsgReq) (*pbChat.SendMsgResp, error) {
	replay := pbChat.SendMsgResp{}
	res, ok, err := c.doBeforeFilters(pb)
	if err != nil {
		return returnMsg(&replay, pb, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), err.Error(), 0)
	}
	if !ok {
		return res, nil
	}

	// fmt.Println("SEND_MSG:before send filters do over")

	res, err = c.rpc.doSendMsg(c.ctx, pb)
	if err != nil {
		return res, err
	}

	res, ok, err = c.doAfterFilters(pb, res)
	if err != nil {
		return returnMsg(&replay, pb, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), err.Error(), 0)
	}
	if !ok {
		return res, nil
	}

	return res, nil
}

func (c *SendContext) SetCtx(ctx context.Context) {
	c.ctx = ctx
}

func (c *SendContext) WithValue(key, val interface{}) {
	ctx := context.WithValue(c.ctx, key, val)
	c.SetCtx(ctx)
}

// Copy returns a copy of the current context that can be safely used outside the request's scope.
// This has to be used when the context has to be passed to a goroutine or wrapped by WithTimeout etc.
func (c *SendContext) Copy() *SendContext {
	cp := SendContext{
		ctx:           c.ctx,
		rpc:           c.rpc,
		beforeFilters: c.beforeFilters,
		afterSenders:  c.afterSenders,
	}

	return &cp
}

/************************************/
/***** context *****/
/************************************/

// Deadline returns that there is no deadline (ok==false) when c has no Context.
func (c *SendContext) Deadline() (deadline time.Time, ok bool) {
	if c.ctx == nil {
		return
	}
	return c.ctx.Deadline()
}

// Done returns nil (chan which will wait forever) when c has no Context.
func (c *SendContext) Done() <-chan struct{} {
	if c.ctx == nil {
		return nil
	}
	return c.ctx.Done()
}

// Err returns nil when c has no Context.
func (c *SendContext) Err() error {
	if c.ctx == nil {
		return nil
	}
	return c.ctx.Err()
}

func (c *SendContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}
