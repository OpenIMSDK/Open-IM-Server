package friend

import (
	chat "Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	cp "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbFriend "Open_IM/pkg/proto/friend"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type friendServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func NewFriendServer(port int) *friendServer {
	log.NewPrivateLog("friend")
	return &friendServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImFriendName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (s *friendServer) Run() {
	log.NewInfo("0", "friendServer run...")

	ip := utils.ServerIP
	registerAddress := ip + ":" + strconv.Itoa(s.rpcPort)
	//listener network
	listener, err := net.Listen("tcp", registerAddress)
	if err != nil {
		log.NewError("0", "Listen failed ", err.Error(), registerAddress)
		return
	}
	log.NewInfo("0", "listen ok ", registerAddress)
	defer listener.Close()
	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	//User friend related services register to etcd
	pbFriend.RegisterFriendServer(srv, s)
	host := viper.GetString("endpoints.rpc_friend")
	err = getcdv3.RegisterEtcd(s.etcdSchema, strings.Join(s.etcdAddr, ","), host, s.rpcPort, s.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error(), s.etcdSchema, strings.Join(s.etcdAddr, ","), ip, s.rpcPort, s.rpcRegisterName)
		return
	}
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error(), listener)
		return
	}
}

func (s *friendServer) AddBlacklist(ctx context.Context, req *pbFriend.AddBlacklistReq) (*pbFriend.AddBlacklistResp, error) {
	log.NewInfo(req.CommID.OperationID, "AddBlacklist args ", req.String())
	ok := token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID)
	if !ok {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.AddBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	black := db.Black{OwnerUserID: req.CommID.FromUserID, BlockUserID: req.CommID.ToUserID, OperatorUserID: req.CommID.OpUserID}

	err := imdb.InsertInToUserBlackList(black)
	if err != nil {
		log.NewError(req.CommID.OperationID, "InsertInToUserBlackList failed ", err.Error())
		return &pbFriend.AddBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	log.NewInfo(req.CommID.OperationID, "AddBlacklist rpc ok ", req.CommID.FromUserID, req.CommID.ToUserID)
	chat.BlackAddedNotification(req)
	return &pbFriend.AddBlacklistResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) AddFriend(ctx context.Context, req *pbFriend.AddFriendReq) (*pbFriend.AddFriendResp, error) {
	log.NewInfo(req.CommID.OperationID, "AddFriend args ", req.String())
	ok := token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID)
	if !ok {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	//Cannot add non-existent users
	if _, err := imdb.GetUserByUserID(req.CommID.ToUserID); err != nil {
		log.NewError(req.CommID.OperationID, "GetUserByUserID failed ", err.Error(), req.CommID.ToUserID)
		return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	//Establish a latest relationship in the friend request table
	friendRequest := db.FriendRequest{
		HandleResult: 0, ReqMsg: req.ReqMsg, CreateTime: time.Now()}
	utils.CopyStructFields(&friendRequest, req.CommID)
	// {openIM001 openIM002 0 test add friend 0001-01-01 00:00:00 +0000 UTC   0001-01-01 00:00:00 +0000 UTC }]
	log.NewDebug(req.CommID.OperationID, "UpdateFriendApplication args ", friendRequest)
	//err := imdb.InsertFriendApplication(&friendRequest)
	err := imdb.InsertFriendApplication(&friendRequest,
		map[string]interface{}{"handle_result": 0, "req_msg": friendRequest.ReqMsg, "create_time": friendRequest.CreateTime,
			"handler_user_id": "", "handle_msg": "", "handle_time": utils.UnixSecondToTime(0), "ex": ""})
	if err != nil {
		log.NewError(req.CommID.OperationID, "UpdateFriendApplication failed ", err.Error(), friendRequest)
		return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	chat.FriendApplicationNotification(req)
	return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) ImportFriend(ctx context.Context, req *pbFriend.ImportFriendReq) (*pbFriend.ImportFriendResp, error) {
	log.NewInfo(req.OperationID, "ImportFriend args ", req.String())
	resp := pbFriend.ImportFriendResp{CommonResp: &pbFriend.CommonResp{}}
	var c pbFriend.CommonResp

	if !utils.IsContain(req.OpUserID, config.Config.Manager.AppManagerUid) {
		log.NewError(req.OperationID, "not authorized", req.OpUserID, config.Config.Manager.AppManagerUid)
		c.ErrCode = constant.ErrAccess.ErrCode
		c.ErrMsg = constant.ErrAccess.ErrMsg
		for _, v := range req.FriendUserIDList {
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		}
		resp.CommonResp = &c
		return &resp, nil
	}
	if _, err := imdb.GetUserByUserID(req.FromUserID); err != nil {
		log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), req.FromUserID)
		c.ErrCode = constant.ErrDB.ErrCode
		c.ErrMsg = "this user not exists,cant not add friend"
		for _, v := range req.FriendUserIDList {
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		}
		resp.CommonResp = &c
		return &resp, nil
	}

	for _, v := range req.FriendUserIDList {
		log.NewDebug(req.OperationID, "FriendUserIDList ", v)
		if _, fErr := imdb.GetUserByUserID(v); fErr != nil {
			log.NewError(req.OperationID, "GetUserByUserID failed ", fErr.Error(), v)
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		} else {
			if _, err := imdb.GetFriendRelationshipFromFriend(req.FromUserID, v); err != nil {
				//Establish two single friendship
				toInsertFollow := db.Friend{OwnerUserID: req.FromUserID, FriendUserID: v}
				err1 := imdb.InsertToFriend(&toInsertFollow)
				if err1 != nil {
					log.NewError(req.OperationID, "InsertToFriend failed ", err1.Error(), toInsertFollow)
					resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
					continue
				}
				toInsertFollow = db.Friend{OwnerUserID: v, FriendUserID: req.FromUserID}
				err2 := imdb.InsertToFriend(&toInsertFollow)
				if err2 != nil {
					log.NewError(req.OperationID, "InsertToFriend failed ", err2.Error(), toInsertFollow)
					resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
					continue
				}
				resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: 0})
				log.NewDebug(req.OperationID, "UserIDResultList ", resp.UserIDResultList)
				chat.FriendAddedNotification(req.OperationID, req.OpUserID, req.FromUserID, v)
			} else {
				log.NewWarn(req.OperationID, "GetFriendRelationshipFromFriend ok", req.FromUserID, v)
				resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: 0})
			}
		}
	}
	resp.CommonResp.ErrCode = 0
	log.NewInfo(req.OperationID, "ImportFriend rpc ok ", resp.String())
	return &resp, nil
}

//process Friend application
func (s *friendServer) AddFriendResponse(ctx context.Context, req *pbFriend.AddFriendResponseReq) (*pbFriend.AddFriendResponseResp, error) {
	log.NewInfo(req.CommID.OperationID, "AddFriendResponse args ", req.String())

	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}

	//Check there application before agreeing or refuse to a friend's application
	//req.CommID.FromUserID process req.CommID.ToUserID
	friendRequest, err := imdb.GetFriendApplicationByBothUserID(req.CommID.ToUserID, req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "GetFriendApplicationByBothUserID failed ", err.Error(), req.CommID.ToUserID, req.CommID.FromUserID)
		return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	friendRequest.HandleResult = req.HandleResult
	friendRequest.HandleTime = time.Now()
	//friendRequest.HandleTime.Unix()
	friendRequest.HandleMsg = req.HandleMsg
	friendRequest.HandlerUserID = req.CommID.OpUserID
	err = imdb.UpdateFriendApplication(friendRequest)
	if err != nil {
		log.NewError(req.CommID.OperationID, "UpdateFriendApplication failed ", err.Error(), friendRequest)
		return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	//Change the status of the friend request form
	if req.HandleResult == constant.FriendFlag {
		//Establish friendship after find friend relationship not exists
		_, err := imdb.GetFriendRelationshipFromFriend(req.CommID.FromUserID, req.CommID.ToUserID)
		if err == nil {
			log.NewWarn(req.CommID.OperationID, "GetFriendRelationshipFromFriend exist", req.CommID.FromUserID, req.CommID.ToUserID)
		} else {
			//Establish two single friendship
			toInsertFollow := db.Friend{OwnerUserID: req.CommID.FromUserID, FriendUserID: req.CommID.ToUserID, OperatorUserID: req.CommID.OpUserID}
			err = imdb.InsertToFriend(&toInsertFollow)
			if err != nil {
				log.NewError(req.CommID.OperationID, "InsertToFriend failed ", err.Error(), toInsertFollow)
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
			}
		}

		_, err = imdb.GetFriendRelationshipFromFriend(req.CommID.ToUserID, req.CommID.FromUserID)
		if err == nil {
			log.NewWarn(req.CommID.OperationID, "GetFriendRelationshipFromFriend exist", req.CommID.ToUserID, req.CommID.FromUserID)
		} else {
			toInsertFollow := db.Friend{OwnerUserID: req.CommID.ToUserID, FriendUserID: req.CommID.FromUserID, OperatorUserID: req.CommID.OpUserID}
			err = imdb.InsertToFriend(&toInsertFollow)
			if err != nil {
				log.NewError(req.CommID.OperationID, "InsertToFriend failed ", err.Error(), toInsertFollow)
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
			}
			chat.FriendAddedNotification(req.CommID.OperationID, req.CommID.OpUserID, req.CommID.FromUserID, req.CommID.ToUserID)
		}
	}
	if req.HandleResult == constant.FriendResponseAgree {
		chat.FriendApplicationApprovedNotification(req)
	} else if req.HandleResult == constant.FriendResponseRefuse {
		chat.FriendApplicationRejectedNotification(req)
	} else {
		log.Error(req.CommID.OperationID, "HandleResult failed ", req.HandleResult)
	}
	log.NewInfo(req.CommID.OperationID, "rpc AddFriendResponse ok")
	return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) DeleteFriend(ctx context.Context, req *pbFriend.DeleteFriendReq) (*pbFriend.DeleteFriendResp, error) {
	log.NewInfo(req.CommID.OperationID, "DeleteFriend args ", req.String())
	//Parse token, to find current user information
	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.DeleteFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	err := imdb.DeleteSingleFriendInfo(req.CommID.FromUserID, req.CommID.ToUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "DeleteSingleFriendInfo failed", err.Error(), req.CommID.FromUserID, req.CommID.ToUserID)
		return &pbFriend.DeleteFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	log.NewInfo(req.CommID.OperationID, "DeleteFriend rpc ok")
	chat.FriendDeletedNotification(req)
	return &pbFriend.DeleteFriendResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) GetBlacklist(ctx context.Context, req *pbFriend.GetBlacklistReq) (*pbFriend.GetBlacklistResp, error) {
	log.NewInfo(req.CommID.OperationID, "GetBlacklist args ", req.String())

	//Parse token, to find current user information
	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.GetBlacklistResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}

	blackListInfo, err := imdb.GetBlackListByUserID(req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "GetBlackListByUID failed ", err.Error(), req.CommID.FromUserID)
		return &pbFriend.GetBlacklistResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}

	var (
		userInfoList []*sdkws.PublicUserInfo
	)
	for _, blackUser := range blackListInfo {
		var blackUserInfo sdkws.PublicUserInfo
		//Find black user information
		us, err := imdb.GetUserByUserID(blackUser.BlockUserID)
		if err != nil {
			log.NewError(req.CommID.OperationID, "GetUserByUserID failed ", err.Error(), blackUser.BlockUserID)
			continue
		}
		utils.CopyStructFields(&blackUserInfo, us)
		userInfoList = append(userInfoList, &blackUserInfo)
	}
	log.NewInfo(req.CommID.OperationID, "rpc GetBlacklist ok ", pbFriend.GetBlacklistResp{BlackUserInfoList: userInfoList})
	return &pbFriend.GetBlacklistResp{BlackUserInfoList: userInfoList}, nil
}

func (s *friendServer) SetFriendRemark(ctx context.Context, req *pbFriend.SetFriendRemarkReq) (*pbFriend.SetFriendRemarkResp, error) {
	log.NewInfo(req.CommID.OperationID, "SetFriendComment args ", req.String())
	//Parse token, to find current user information
	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.SetFriendRemarkResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}

	err := imdb.UpdateFriendComment(req.CommID.FromUserID, req.CommID.ToUserID, req.Remark)
	if err != nil {
		log.NewError(req.CommID.OperationID, "UpdateFriendComment failed ", req.CommID.FromUserID, req.CommID.ToUserID, req.Remark)
		return &pbFriend.SetFriendRemarkResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	log.NewInfo(req.CommID.OperationID, "rpc SetFriendComment ok")
	chat.FriendRemarkSetNotification(req.CommID.OperationID, req.CommID.OpUserID, req.CommID.FromUserID, req.CommID.ToUserID)
	return &pbFriend.SetFriendRemarkResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) RemoveBlacklist(ctx context.Context, req *pbFriend.RemoveBlacklistReq) (*pbFriend.RemoveBlacklistResp, error) {
	log.NewInfo(req.CommID.OperationID, "RemoveBlacklist args ", req.String())
	//Parse token, to find current user information
	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.RemoveBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	err := imdb.RemoveBlackList(req.CommID.FromUserID, req.CommID.ToUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "RemoveBlackList failed", err.Error(), req.CommID.FromUserID, req.CommID.ToUserID)
		return &pbFriend.RemoveBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil

	}
	log.NewInfo(req.CommID.OperationID, "rpc RemoveBlacklist ok ")
	chat.BlackDeletedNotification(req)
	return &pbFriend.RemoveBlacklistResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) IsInBlackList(ctx context.Context, req *pbFriend.IsInBlackListReq) (*pbFriend.IsInBlackListResp, error) {
	log.NewInfo("IsInBlackList args ", req.String())
	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.IsInBlackListResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}

	var isInBlacklist = false
	err := imdb.CheckBlack(req.CommID.FromUserID, req.CommID.ToUserID)
	if err == nil {
		isInBlacklist = true
	}
	log.NewInfo(req.CommID.OperationID, "IsInBlackList rpc ok ", pbFriend.IsInBlackListResp{Response: isInBlacklist})
	return &pbFriend.IsInBlackListResp{Response: isInBlacklist}, nil
}

func (s *friendServer) IsFriend(ctx context.Context, req *pbFriend.IsFriendReq) (*pbFriend.IsFriendResp, error) {
	log.NewInfo("IsFriend args ", req.String())
	var isFriend bool
	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.IsFriendResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}
	_, err := imdb.GetFriendRelationshipFromFriend(req.CommID.FromUserID, req.CommID.ToUserID)
	if err == nil {
		isFriend = true
	} else {
		isFriend = false
	}
	log.NewInfo("IsFriend rpc ok ", pbFriend.IsFriendResp{Response: isFriend})
	return &pbFriend.IsFriendResp{Response: isFriend}, nil
}

func (s *friendServer) GetFriendList(ctx context.Context, req *pbFriend.GetFriendListReq) (*pbFriend.GetFriendListResp, error) {
	log.NewInfo("GetFriendList args ", req.String())
	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.GetFriendListResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}

	friends, err := imdb.GetFriendListByUserID(req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "FindUserInfoFromFriend failed ", err.Error(), req.CommID.FromUserID)
		return &pbFriend.GetFriendListResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	var userInfoList []*sdkws.FriendInfo
	for _, friendUser := range friends {
		friendUserInfo := sdkws.FriendInfo{FriendUser: &sdkws.UserInfo{}}
		cp.FriendDBCopyOpenIM(&friendUserInfo, &friendUser)
		log.NewDebug(req.CommID.OperationID, "friends : ", friendUser, "openim friends: ", friendUserInfo)
		userInfoList = append(userInfoList, &friendUserInfo)
	}
	log.NewInfo(req.CommID.OperationID, "rpc GetFriendList ok", pbFriend.GetFriendListResp{FriendInfoList: userInfoList})
	return &pbFriend.GetFriendListResp{FriendInfoList: userInfoList}, nil
}

//received
func (s *friendServer) GetFriendApplyList(ctx context.Context, req *pbFriend.GetFriendApplyListReq) (*pbFriend.GetFriendApplyListResp, error) {
	log.NewInfo(req.CommID.OperationID, "GetFriendApplyList args ", req.String())
	//Parse token, to find current user information
	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.GetFriendApplyListResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}
	//	Find the  current user friend applications received
	ApplyUsersInfo, err := imdb.GetReceivedFriendsApplicationListByUserID(req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "GetReceivedFriendsApplicationListByUserID ", err.Error(), req.CommID.FromUserID)
		return &pbFriend.GetFriendApplyListResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}

	var appleUserList []*sdkws.FriendRequest
	for _, applyUserInfo := range ApplyUsersInfo {
		var userInfo sdkws.FriendRequest
		cp.FriendRequestDBCopyOpenIM(&userInfo, &applyUserInfo)
		//	utils.CopyStructFields(&userInfo, applyUserInfo)
		//	u, err := imdb.GetUserByUserID(userInfo.FromUserID)
		//	if err != nil {
		//		log.Error(req.CommID.OperationID, "GetUserByUserID", userInfo.FromUserID)
		//		continue
		//	}
		//	userInfo.FromNickname = u.Nickname
		//	userInfo.FromFaceURL = u.FaceURL
		//	userInfo.FromGender = u.Gender
		//
		//	u, err = imdb.GetUserByUserID(userInfo.ToUserID)
		//	if err != nil {
		//		log.Error(req.CommID.OperationID, "GetUserByUserID", userInfo.ToUserID)
		//		continue
		//	}
		//	userInfo.ToNickname = u.Nickname
		//	userInfo.ToFaceURL = u.FaceURL
		//	userInfo.ToGender = u.Gender
		appleUserList = append(appleUserList, &userInfo)
	}

	log.NewInfo(req.CommID.OperationID, "rpc GetFriendApplyList ok", pbFriend.GetFriendApplyListResp{FriendRequestList: appleUserList})
	return &pbFriend.GetFriendApplyListResp{FriendRequestList: appleUserList}, nil
}

func (s *friendServer) GetSelfApplyList(ctx context.Context, req *pbFriend.GetSelfApplyListReq) (*pbFriend.GetSelfApplyListResp, error) {
	log.NewInfo(req.CommID.OperationID, "GetSelfApplyList args ", req.String())

	//Parse token, to find current user information
	if !token_verify.CheckAccess(req.CommID.OpUserID, req.CommID.FromUserID) {
		log.NewError(req.CommID.OperationID, "CheckAccess false ", req.CommID.OpUserID, req.CommID.FromUserID)
		return &pbFriend.GetSelfApplyListResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}
	//	Find the self add other userinfo
	usersInfo, err := imdb.GetSendFriendApplicationListByUserID(req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "GetSendFriendApplicationListByUserID failed ", err.Error(), req.CommID.FromUserID)
		return &pbFriend.GetSelfApplyListResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	var selfApplyOtherUserList []*sdkws.FriendRequest
	for _, selfApplyOtherUserInfo := range usersInfo {
		var userInfo sdkws.FriendRequest // pbFriend.ApplyUserInfo
		cp.FriendRequestDBCopyOpenIM(&userInfo, &selfApplyOtherUserInfo)
		//u, err := imdb.GetUserByUserID(userInfo.FromUserID)
		//if err != nil {
		//	log.Error(req.CommID.OperationID, "GetUserByUserID", userInfo.FromUserID)
		//	continue
		//}
		//userInfo.FromNickname = u.Nickname
		//userInfo.FromFaceURL = u.FaceURL
		//userInfo.FromGender = u.Gender
		//
		//u, err = imdb.GetUserByUserID(userInfo.ToUserID)
		//if err != nil {
		//	log.Error(req.CommID.OperationID, "GetUserByUserID", userInfo.ToUserID)
		//	continue
		//}
		//userInfo.ToNickname = u.Nickname
		//userInfo.ToFaceURL = u.FaceURL
		//userInfo.ToGender = u.Gender

		selfApplyOtherUserList = append(selfApplyOtherUserList, &userInfo)
	}
	log.NewInfo(req.CommID.OperationID, "rpc GetSelfApplyList ok", pbFriend.GetSelfApplyListResp{FriendRequestList: selfApplyOtherUserList})
	return &pbFriend.GetSelfApplyListResp{FriendRequestList: selfApplyOtherUserList}, nil
}

////
//func (s *friendServer) GetFriendsInfo(ctx context.Context, req *pbFriend.GetFriendsInfoReq) (*pbFriend.GetFriendInfoResp, error) {
//	return nil, nil
////	log.NewInfo(req.CommID.OperationID, "GetFriendsInfo args ", req.String())
////	var (
////		isInBlackList int32
////		//	isFriend      int32
////		comment string
////	)
////
////	friendShip, err := imdb.FindFriendRelationshipFromFriend(req.CommID.FromUserID, req.CommID.ToUserID)
////	if err != nil {
////		log.NewError(req.CommID.OperationID, "FindFriendRelationshipFromFriend failed ", err.Error())
////		return &pbFriend.GetFriendInfoResp{ErrCode: constant.ErrSearchUserInfo.ErrCode, ErrMsg: constant.ErrSearchUserInfo.ErrMsg}, nil
////		//	isFriend = constant.FriendFlag
////	}
////	comment = friendShip.Remark
////
////	friendUserInfo, err := imdb.FindUserByUID(req.CommID.ToUserID)
////	if err != nil {
////		log.NewError(req.CommID.OperationID, "FindUserByUID failed ", err.Error())
////		return &pbFriend.GetFriendInfoResp{ErrCode: constant.ErrSearchUserInfo.ErrCode, ErrMsg: constant.ErrSearchUserInfo.ErrMsg}, nil
////	}
////
////	err = imdb.FindRelationshipFromBlackList(req.CommID.FromUserID, req.CommID.ToUserID)
////	if err == nil {
////		isInBlackList = constant.BlackListFlag
////	}
////
////	resp := pbFriend.GetFriendInfoResp{ErrCode: 0, ErrMsg:  "",}
////
////	utils.CopyStructFields(resp.FriendInfoList, friendUserInfo)
////	resp.Data.IsBlack = isInBlackList
////	resp.Data.OwnerUserID = req.CommID.FromUserID
////	resp.Data.Remark = comment
////	resp.Data.CreateTime = friendUserInfo.CreateTime
////
////	log.NewInfo(req.CommID.OperationID, "GetFriendsInfo ok ", resp)
////	return &resp, nil
////
//}
