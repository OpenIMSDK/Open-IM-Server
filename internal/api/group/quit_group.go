package group

import (
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pb "Open_IM/pkg/proto/group"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// paramsQuitGroup struct
type paramsQuitGroup struct {
	GroupID     string `json:"groupID" binding:"required"`
	OperationID string `json:"operationID" binding:"required"`
}

// @Summary
// @Schemes
// @Description quit group
// @Tags group
// @Accept json
// @Produce json
// @Param body body group.paramsQuitGroup true "quit group"
// @Param token header string true "token"
// @Success 200 {object} user.result
// @Failure 400 {object} user.result
// @Failure 500 {object} user.result
// @Router /group/set_group_info [post]
func QuitGroup(c *gin.Context) {
	log.Info("", "", "api quit group init ....")

	etcdConn := getcdv3.GetGroupConn()
	client := pb.NewGroupClient(etcdConn)
	//defer etcdConn.Close()

	params := paramsQuitGroup{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &pb.QuitGroupReq{
		GroupID:     params.GroupID,
		OperationID: params.OperationID,
		Token:       c.Request.Header.Get("token"),
	}
	log.Info(req.Token, req.OperationID, "api quit group is server,params=%s", req.String())
	RpcResp, err := client.QuitGroup(context.Background(), req)
	if err != nil {
		log.Error(req.Token, req.OperationID, "call quit group rpc server failed,err=%s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": "call  rpc server failed"})
		return
	}
	log.InfoByArgs("call quit group rpc server success,args=%s", RpcResp.String())
	c.JSON(http.StatusOK, gin.H{"errCode": RpcResp.ErrorCode, "errMsg": RpcResp.ErrorMsg})
	log.InfoByArgs("api quit group success return,get args=%s,return args=%s", req.String(), RpcResp.String())
}