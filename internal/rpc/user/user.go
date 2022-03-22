package user

import (
	chat "Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	errors "Open_IM/pkg/common/http"

	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbFriend "Open_IM/pkg/proto/friend"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	pbUser "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type userServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func NewUserServer(port int) *userServer {
	log.NewPrivateLog("user")
	return &userServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImUserName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (s *userServer) Run() {
	log.NewInfo("0", "", "rpc user start...")

	ip := utils.ServerIP
	registerAddress := ip + ":" + strconv.Itoa(s.rpcPort)
	//listener network
	listener, err := net.Listen("tcp", registerAddress)
	if err != nil {
		log.NewError("0", "listen network failed ", err.Error(), registerAddress)
		return
	}
	log.NewInfo("0", "listen network success, address ", registerAddress, listener)
	defer listener.Close()
	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	//Service registers with etcd
	pbUser.RegisterUserServer(srv, s)
	host := viper.GetString("endpoints.rpc_user")
	err = getcdv3.RegisterEtcd(s.etcdSchema, strings.Join(s.etcdAddr, ","), host, s.rpcPort, s.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error(), s.etcdSchema, strings.Join(s.etcdAddr, ","), ip, s.rpcPort, s.rpcRegisterName)
		return
	}
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("0", "rpc  user success")
}

func (s *userServer) GetUserInfo(ctx context.Context, req *pbUser.GetUserInfoReq) (*pbUser.GetUserInfoResp, error) {
	log.NewInfo(req.OperationID, "GetUserInfo args ", req.String())
	var userInfoList []*sdkws.UserInfo
	if len(req.UserIDList) > 0 {
		for _, userID := range req.UserIDList {
			var userInfo sdkws.UserInfo
			user, err := imdb.GetUserByUserID(userID)
			if err != nil {
				log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), userID)
				continue
			}
			utils.CopyStructFields(&userInfo, user)
			userInfo.Birth = uint32(user.Birth.Unix())
			userInfoList = append(userInfoList, &userInfo)
		}
	} else {
		return &pbUser.GetUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}
	log.NewInfo(req.OperationID, "GetUserInfo rpc return ", pbUser.GetUserInfoResp{CommonResp: &pbUser.CommonResp{}, UserInfoList: userInfoList})
	return &pbUser.GetUserInfoResp{CommonResp: &pbUser.CommonResp{}, UserInfoList: userInfoList}, nil
}

func (s *userServer) BatchSetConversations(ctx context.Context, req *pbUser.BatchSetConversationsReq) (*pbUser.BatchSetConversationsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.BatchSetConversationsResp{}
	for _, v := range req.Conversations {
		conversation := db.Conversation{}
		if err := utils.CopyStructFields(&conversation, v); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), v.String(), "CopyStructFields failed", err.Error())
		}
		if err := db.DB.SetSingleConversationRecvMsgOpt(req.OwnerUserID, v.ConversationID, v.RecvMsgOpt); err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, rpc return", err.Error())
			resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
			return resp, nil
		}
		if err := imdb.SetConversation(conversation); err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
			resp.Failed = append(resp.Failed, v.ConversationID)
			continue
		}
		resp.Success = append(resp.Success, v.ConversationID)
	}
	chat.SetConversationNotification(req.OperationID, req.OwnerUserID)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) GetAllConversations(ctx context.Context, req *pbUser.GetAllConversationsReq) (*pbUser.GetAllConversationsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetAllConversationsResp{Conversations: []*pbUser.Conversation{}}
	conversations, err := imdb.GetUserAllConversations(req.OwnerUserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetConversations error", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	if err = utils.CopyStructFields(&resp.Conversations, conversations); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields error", err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) GetConversation(ctx context.Context, req *pbUser.GetConversationReq) (*pbUser.GetConversationResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetConversationResp{Conversation: &pbUser.Conversation{}}
	conversation, err := imdb.GetConversation(req.OwnerUserID, req.ConversationID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetConversation error", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	if err := utils.CopyStructFields(resp.Conversation, &conversation); err != nil {
		log.Debug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields error", conversation, err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) GetConversations(ctx context.Context, req *pbUser.GetConversationsReq) (*pbUser.GetConversationsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetConversationsResp{Conversations: []*pbUser.Conversation{}}
	conversations, err := imdb.GetConversations(req.OwnerUserID, req.ConversationIDs)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetConversations error", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	if err := utils.CopyStructFields(&resp.Conversations, conversations); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", conversations, err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) SetConversation(ctx context.Context, req *pbUser.SetConversationReq) (*pbUser.SetConversationResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.SetConversationResp{}
	var conversation db.Conversation
	if err := utils.CopyStructFields(&conversation, req.Conversation); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", *req.Conversation, err.Error())
	}
	if err := db.DB.SetSingleConversationRecvMsgOpt(req.Conversation.OwnerUserID, req.Conversation.ConversationID, req.Conversation.RecvMsgOpt); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, rpc return", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	err := imdb.SetConversation(conversation)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	chat.SetConversationNotification(req.OperationID, req.Conversation.OwnerUserID)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) SetRecvMsgOpt(ctx context.Context, req *pbUser.SetRecvMsgOptReq) (*pbUser.SetRecvMsgOptResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.SetRecvMsgOptResp{}
	var conversation db.Conversation
	if err := utils.CopyStructFields(&conversation, req); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", *req, err.Error())
	}
	if err := db.DB.SetSingleConversationRecvMsgOpt(req.OwnerUserID, req.ConversationID, req.RecvMsgOpt); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, rpc return", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	err := imdb.SetRecvMsgOpt(conversation)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	chat.SetConversationNotification(req.OperationID, req.OwnerUserID)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) DeleteUsers(_ context.Context, req *pbUser.DeleteUsersReq) (*pbUser.DeleteUsersResp, error) {
	log.NewInfo(req.OperationID, "DeleteUsers args ", req.String())
	if !token_verify.IsMangerUserID(req.OpUserID) {
		log.NewError(req.OperationID, "IsMangerUserID false ", req.OpUserID)
		return &pbUser.DeleteUsersResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, FailedUserIDList: req.DeleteUserIDList}, nil
	}
	var common pbUser.CommonResp
	resp := pbUser.DeleteUsersResp{CommonResp: &common}
	for _, userID := range req.DeleteUserIDList {
		i := imdb.DeleteUser(userID)
		if i == 0 {
			log.NewError(req.OperationID, "delete user error", userID)
			common.ErrCode = 201
			common.ErrMsg = "some uid deleted failed"
			resp.FailedUserIDList = append(resp.FailedUserIDList, userID)
		}
	}
	log.NewInfo(req.OperationID, "DeleteUsers rpc return ", resp.String())
	return &resp, nil
}

func (s *userServer) GetAllUserID(_ context.Context, req *pbUser.GetAllUserIDReq) (*pbUser.GetAllUserIDResp, error) {
	log.NewInfo(req.OperationID, "GetAllUserID args ", req.String())
	if !token_verify.IsMangerUserID(req.OpUserID) {
		log.NewError(req.OperationID, "IsMangerUserID false ", req.OpUserID)
		return &pbUser.GetAllUserIDResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	uidList, err := imdb.SelectAllUserID()
	if err != nil {
		log.NewError(req.OperationID, "SelectAllUserID false ", err.Error())
		return &pbUser.GetAllUserIDResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	} else {
		log.NewInfo(req.OperationID, "GetAllUserID rpc return ", pbUser.GetAllUserIDResp{CommonResp: &pbUser.CommonResp{}, UserIDList: uidList})
		return &pbUser.GetAllUserIDResp{CommonResp: &pbUser.CommonResp{}, UserIDList: uidList}, nil
	}
}

func (s *userServer) AccountCheck(_ context.Context, req *pbUser.AccountCheckReq) (*pbUser.AccountCheckResp, error) {
	log.NewInfo(req.OperationID, "AccountCheck args ", req.String())
	if !token_verify.IsMangerUserID(req.OpUserID) {
		log.NewError(req.OperationID, "IsMangerUserID false ", req.OpUserID)
		return &pbUser.AccountCheckResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	uidList, err := imdb.SelectSomeUserID(req.CheckUserIDList)
	log.NewDebug(req.OperationID, "from db uid list is:", uidList)
	if err != nil {
		log.NewError(req.OperationID, "SelectSomeUserID failed ", err.Error(), req.CheckUserIDList)
		return &pbUser.AccountCheckResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	} else {
		var r []*pbUser.AccountCheckResp_SingleUserStatus
		for _, v := range req.CheckUserIDList {
			temp := new(pbUser.AccountCheckResp_SingleUserStatus)
			temp.UserID = v
			if utils.IsContain(v, uidList) {
				temp.AccountStatus = constant.Registered
			} else {
				temp.AccountStatus = constant.UnRegistered
			}
			r = append(r, temp)
		}
		resp := pbUser.AccountCheckResp{CommonResp: &pbUser.CommonResp{ErrCode: 0, ErrMsg: ""}, ResultList: r}
		log.NewInfo(req.OperationID, "AccountCheck rpc return ", resp.String())
		return &resp, nil
	}

}

func (s *userServer) UpdateUserInfo(ctx context.Context, req *pbUser.UpdateUserInfoReq) (*pbUser.UpdateUserInfoResp, error) {
	log.NewInfo(req.OperationID, "UpdateUserInfo args ", req.String())
	if !token_verify.CheckAccess(req.OpUserID, req.UserInfo.UserID) {
		log.NewError(req.OperationID, "CheckAccess false ", req.OpUserID, req.UserInfo.UserID)
		return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}

	var user db.User
	utils.CopyStructFields(&user, req.UserInfo)
	if req.UserInfo.Birth != 0 {
		user.Birth = utils.UnixSecondToTime(int64(req.UserInfo.Birth))
	}
	err := imdb.UpdateUserInfo(user)
	if err != nil {
		log.NewError(req.OperationID, "UpdateUserInfo failed ", err.Error(), user)
		return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName)
	client := pbFriend.NewFriendClient(etcdConn)
	newReq := &pbFriend.GetFriendListReq{
		CommID: &pbFriend.CommID{OperationID: req.OperationID, FromUserID: req.UserInfo.UserID, OpUserID: req.OpUserID},
	}

	RpcResp, err := client.GetFriendList(context.Background(), newReq)
	if err != nil {
		log.NewError(req.OperationID, "GetFriendList failed ", err.Error(), newReq)
		return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{}}, nil
	}
	for _, v := range RpcResp.FriendInfoList {
		log.Info(req.OperationID, "UserInfoUpdatedNotification ", req.UserInfo.UserID, v.FriendUser.UserID)
		chat.UserInfoUpdatedNotification(req.OperationID, req.UserInfo.UserID, v.FriendUser.UserID)
	}
	chat.UserInfoUpdatedNotification(req.OperationID, req.UserInfo.UserID, req.OpUserID)
	log.Info(req.OperationID, "UserInfoUpdatedNotification ", req.UserInfo.UserID, req.OpUserID)
	return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{}}, nil
}

func (s *userServer) GetUsersByName(ctx context.Context, req *pbUser.GetUsersByNameReq) (*pbUser.GetUsersByNameResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbUser.GetUsersByNameResp{}
	users, err := imdb.GetUserByName(req.UserName, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserByName failed", err.Error())
		return resp, errors.WrapError(constant.ErrDB)
	}
	for _, user := range users {
		isBlock, err := imdb.UserIsBlock(user.UserID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
			continue
		}
		resp.Users = append(resp.Users, &pbUser.User{
			ProfilePhoto: user.FaceURL,
			Nickname:     user.Nickname,
			UserId:       user.UserID,
			CreateTime:   user.CreateTime.String(),
			IsBlock:      isBlock,
		})
	}
	user := db.User{Nickname: req.UserName}
	userNums, err := imdb.GetUsersCount(user)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "", err.Error())
		return resp, errors.WrapError(constant.ErrDB)
	}
	resp.UserNums = userNums
	resp.Pagination = &sdkws.ResponsePagination{
		CurrentPage: req.Pagination.PageNumber,
		ShowNumber:  req.Pagination.ShowNumber,
	}
	return resp, nil
}

func (s *userServer) GetUserById(ctx context.Context, req *pbUser.GetUserByIdReq) (*pbUser.GetUserByIdResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbUser.GetUserByIdResp{User: &pbUser.User{}}
	user, err := imdb.GetUserByUserID(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
		return resp, errors.WrapError(constant.ErrDB)
	}
	isBlock, err := imdb.UserIsBlock(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "req：", req.String())
		return resp, errors.WrapError(constant.ErrDB)
	}
	resp.User = &pbUser.User{
		ProfilePhoto: user.FaceURL,
		Nickname:     user.Nickname,
		UserId:       user.UserID,
		CreateTime:   user.CreateTime.String(),
		IsBlock:      isBlock,
	}
	return resp, nil
}

func (s *userServer) GetUsers(ctx context.Context, req *pbUser.GetUsersReq) (*pbUser.GetUsersResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetUsersResp{User: []*pbUser.User{}}
	users, err := imdb.GetUsers(req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		return resp, errors.WrapError(constant.ErrDB)
	}
	for _, v := range users {
		isBlock, err := imdb.UserIsBlock(v.UserID)
		if err == nil {
			user := &pbUser.User{
				ProfilePhoto: v.FaceURL,
				UserId:       v.UserID,
				CreateTime:   v.CreateTime.String(),
				Nickname:     v.Nickname,
				IsBlock:      isBlock,
			}
			resp.User = append(resp.User, user)
		}
	}
	user := db.User{}
	nums, err := imdb.GetUsersCount(user)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsersCount failed", err.Error())
		return resp, errors.WrapError(constant.ErrDB)
	}
	resp.UserNums = nums
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	return resp, nil
}

func (s *userServer) ResignUser(ctx context.Context, req *pbUser.ResignUserReq) (*pbUser.ResignUserResp, error) {
	log.NewInfo(req.OperationID, "ResignUser args ", req.String())
	return &pbUser.ResignUserResp{}, nil
}

func (s *userServer) AlterUser(ctx context.Context, req *pbUser.AlterUserReq) (*pbUser.AlterUserResp, error) {
	log.NewInfo(req.OperationID, "AlterUser args ", req.String())
	resp := &pbUser.AlterUserResp{}
	user := db.User{
		PhoneNumber: strconv.FormatInt(req.PhoneNumber, 10),
		Nickname:    req.Nickname,
		Email:       req.Email,
		UserID:      req.UserId,
	}
	if err := imdb.UpdateUserInfo(user); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "UpdateUserInfo", err.Error())
		return resp, errors.WrapError(constant.ErrDB)
	}
	chat.UserInfoUpdatedNotification(req.OperationID, req.UserId, req.OpUserId)
	return resp, nil
}

func (s *userServer) AddUser(ctx context.Context, req *pbUser.AddUserReq) (*pbUser.AddUserResp, error) {
	log.NewInfo(req.OperationID, "AddUser args ", req.String())
	resp := &pbUser.AddUserResp{}
	err := imdb.AddUser(req.UserId, req.PhoneNumber, req.Name)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser", err.Error())
		return resp, errors.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *userServer) BlockUser(ctx context.Context, req *pbUser.BlockUserReq) (*pbUser.BlockUserResp, error) {
	log.NewInfo(req.OperationID, "BlockUser args ", req.String())
	fmt.Println("BlockUser args ", req.String())
	resp := &pbUser.BlockUserResp{}
	err := imdb.BlockUser(req.UserId, req.EndDisableTime)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BlockUser", err.Error())
		return resp, errors.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *userServer) UnBlockUser(ctx context.Context, req *pbUser.UnBlockUserReq) (*pbUser.UnBlockUserResp, error) {
	log.NewInfo(req.OperationID, "UnBlockUser args ", req.String())
	resp := &pbUser.UnBlockUserResp{}
	err := imdb.UnBlockUser(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "unBlockUser", err.Error())
		return resp, errors.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *userServer) GetBlockUsers(ctx context.Context, req *pbUser.GetBlockUsersReq) (*pbUser.GetBlockUsersResp, error) {
	log.NewInfo(req.OperationID, "GetBlockUsers args ", req.String())
	resp := &pbUser.GetBlockUsersResp{}
	blockUsers, err := imdb.GetBlockUsers(req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.Error(req.OperationID, utils.GetSelfFuncName(), "GetBlockUsers", err.Error())
		return resp, errors.WrapError(constant.ErrDB)
	}
	for _, v := range blockUsers {
		resp.BlockUsers = append(resp.BlockUsers, &pbUser.BlockUser{
			User: &pbUser.User{
				ProfilePhoto: v.User.FaceURL,
				Nickname:     v.User.Nickname,
				UserId:       v.User.UserID,
				IsBlock:      true,
			},
			BeginDisableTime: (v.BeginDisableTime).String(),
			EndDisableTime:   (v.EndDisableTime).String(),
		})
	}
	resp.Pagination = &sdkws.ResponsePagination{}
	resp.Pagination.ShowNumber = req.Pagination.ShowNumber
	resp.Pagination.CurrentPage = req.Pagination.PageNumber
	nums, err := imdb.GetBlockUsersNumCount()
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetBlockUsersNumCount failed", err.Error())
		return resp, errors.WrapError(constant.ErrDB)
	}
	resp.UserNums = nums
	return resp, nil
}

func (s *userServer) GetBlockUserById(_ context.Context, req *pbUser.GetBlockUserByIdReq) (*pbUser.GetBlockUserByIdResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "GetBlockUserById args ", req.String())
	resp := &pbUser.GetBlockUserByIdResp{}
	user, err := imdb.GetBlockUserById(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetBlockUserById", err)
		return resp, errors.WrapError(constant.ErrDB)
	}
	resp.BlockUser = &pbUser.BlockUser{
		User: &pbUser.User{
			ProfilePhoto: user.User.FaceURL,
			Nickname:     user.User.Nickname,
			UserId:       user.User.UserID,
			IsBlock:      true,
		},
		BeginDisableTime: (user.BeginDisableTime).String(),
		EndDisableTime:   (user.EndDisableTime).String(),
	}
	return resp, nil
}

func (s *userServer) DeleteUser(_ context.Context, req *pbUser.DeleteUserReq) (*pbUser.DeleteUserResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())
	resp := &pbUser.DeleteUserResp{}
	if row := imdb.DeleteUser(req.UserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, errors.WrapError(constant.ErrDB)
	}
	return resp, nil
}
