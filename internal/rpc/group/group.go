package group

import (
	chat "Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	cp "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbGroup "Open_IM/pkg/proto/group"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type groupServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func NewGroupServer(port int) *groupServer {
	log.NewPrivateLog("group")
	return &groupServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImGroupName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (s *groupServer) Run() {
	log.NewInfo("0", "group rpc start ")
	ip := utils.ServerIP
	registerAddress := ip + ":" + strconv.Itoa(s.rpcPort)
	//listener network
	listener, err := net.Listen("tcp", registerAddress)
	if err != nil {
		log.NewError("0", "Listen failed ", err.Error(), registerAddress)
		return
	}
	log.NewInfo("0", "listen network success, ", registerAddress, listener)
	defer listener.Close()
	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	//Service registers with etcd
	pbGroup.RegisterGroupServer(srv, s)
	host := viper.GetString("endpoints.rpc_group")
	err = getcdv3.RegisterEtcd(s.etcdSchema, strings.Join(s.etcdAddr, ","), host, s.rpcPort, s.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error())
		return
	}
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("0", "group rpc success")
}

func (s *groupServer) CreateGroup(ctx context.Context, req *pbGroup.CreateGroupReq) (*pbGroup.CreateGroupResp, error) {
	log.NewInfo(req.OperationID, "CreateGroup, args ", req.String())
	if !token_verify.CheckAccess(req.OpUserID, req.OwnerUserID) {
		log.NewError(req.OperationID, "CheckAccess false ", req.OpUserID, req.OwnerUserID)
		return &pbGroup.CreateGroupResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}

	//Time stamp + MD5 to generate group chat id
	groupId := utils.Md5(strconv.FormatInt(time.Now().UnixNano(), 10))
	//to group
	groupInfo := db.Group{}
	utils.CopyStructFields(&groupInfo, req.GroupInfo)
	groupInfo.CreatorUserID = req.OpUserID
	groupInfo.GroupID = groupId
	err := imdb.InsertIntoGroup(groupInfo)
	if err != nil {
		log.NewError(req.OperationID, "InsertIntoGroup failed, ", err.Error(), groupInfo)
		return &pbGroup.CreateGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, http.WrapError(constant.ErrDB)
	}

	us, err := imdb.GetUserByUserID(req.OwnerUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), req.OwnerUserID)
		return &pbGroup.CreateGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, http.WrapError(constant.ErrDB)
	}

	//to group member
	groupMember := db.GroupMember{GroupID: groupId, RoleLevel: constant.GroupOwner, OperatorUserID: req.OpUserID}
	utils.CopyStructFields(&groupMember, us)
	err = imdb.InsertIntoGroupMember(groupMember)
	if err != nil {
		log.NewError(req.OperationID, "InsertIntoGroupMember failed ", err.Error(), groupMember)
		return &pbGroup.CreateGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, http.WrapError(constant.ErrDB)
	}

	err = db.DB.AddGroupMember(groupId, req.OwnerUserID)
	if err != nil {
		log.NewError(req.OperationID, "AddGroupMember failed ", err.Error(), groupId, req.OwnerUserID)
	}
	var okUserIDList []string
	//to group member
	for _, user := range req.InitMemberList {
		us, err := imdb.GetUserByUserID(user.UserID)
		if err != nil {
			log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), user.UserID)
			continue
		}
		if user.RoleLevel == constant.GroupOwner {
			log.NewError(req.OperationID, "only one owner, failed ", user)
			continue
		}
		groupMember.RoleLevel = user.RoleLevel
		utils.CopyStructFields(&groupMember, us)
		err = imdb.InsertIntoGroupMember(groupMember)
		if err != nil {
			log.NewError(req.OperationID, "InsertIntoGroupMember failed ", err.Error(), groupMember)
			continue
		}

		okUserIDList = append(okUserIDList, user.UserID)
		err = db.DB.AddGroupMember(groupId, user.UserID)
		if err != nil {
			log.NewError(req.OperationID, "add mongo group member failed, db.DB.AddGroupMember failed ", err.Error())
		}
	}

	resp := &pbGroup.CreateGroupResp{GroupInfo: &open_im_sdk.GroupInfo{}}
	group, err := imdb.GetGroupInfoByGroupID(groupId)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), groupId)
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		return resp, nil
	}
	utils.CopyStructFields(resp.GroupInfo, group)
	resp.GroupInfo.MemberCount, err = imdb.GetGroupMemberNumByGroupID(groupId)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberNumByGroupID failed ", err.Error(), groupId)
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		return resp, nil
	}
	resp.GroupInfo.OwnerUserID = req.OwnerUserID

	log.NewInfo(req.OperationID, "rpc CreateGroup return ", resp.String())
	chat.GroupCreatedNotification(req.OperationID, req.OpUserID, groupId, okUserIDList)
	return resp, nil
}

func (s *groupServer) GetJoinedGroupList(ctx context.Context, req *pbGroup.GetJoinedGroupListReq) (*pbGroup.GetJoinedGroupListResp, error) {
	log.NewInfo(req.OperationID, "GetJoinedGroupList, args ", req.String())
	if !token_verify.CheckAccess(req.OpUserID, req.FromUserID) {
		log.NewError(req.OperationID, "CheckAccess false ", req.OpUserID, req.FromUserID)
		return &pbGroup.GetJoinedGroupListResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}
	//group list
	joinedGroupList, err := imdb.GetJoinedGroupIDListByUserID(req.FromUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetJoinedGroupIDListByUserID failed ", err.Error(), req.FromUserID)
		return &pbGroup.GetJoinedGroupListResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}

	var resp pbGroup.GetJoinedGroupListResp
	for _, v := range joinedGroupList {
		var groupNode open_im_sdk.GroupInfo
		num, err := imdb.GetGroupMemberNumByGroupID(v)
		owner, err2 := imdb.GetGroupOwnerInfoByGroupID(v)
		group, err := imdb.GetGroupInfoByGroupID(v)
		if num > 0 && owner != nil && err2 == nil && group != nil && err == nil {
			utils.CopyStructFields(&groupNode, group)
			groupNode.CreateTime = uint32(group.CreateTime.Unix())
			groupNode.MemberCount = uint32(num)
			groupNode.OwnerUserID = owner.UserID
			resp.GroupList = append(resp.GroupList, &groupNode)
		} else {
			log.NewError(req.OperationID, "check nil ", num, owner, err, group)
			continue
		}
		log.NewDebug(req.OperationID, "joinedGroup ", groupNode)
	}
	log.NewInfo(req.OperationID, "GetJoinedGroupList rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) InviteUserToGroup(ctx context.Context, req *pbGroup.InviteUserToGroupReq) (*pbGroup.InviteUserToGroupResp, error) {
	log.NewInfo(req.OperationID, "InviteUserToGroup args ", req.String())

	if !imdb.IsExistGroupMember(req.GroupID, req.OpUserID) && !token_verify.IsMangerUserID(req.OpUserID) {
		log.NewError(req.OperationID, "no permission InviteUserToGroup ", req.GroupID, req.OpUserID)
		return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}

	_, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", req.GroupID, err)
		return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}
	//
	//from User:  invite: applicant
	//to user:  invite: invited
	var resp pbGroup.InviteUserToGroupResp
	var okUserIDList []string
	for _, v := range req.InvitedUserIDList {
		var resultNode pbGroup.Id2Result
		resultNode.UserID = v
		resultNode.Result = 0
		toUserInfo, err := imdb.GetUserByUserID(v)
		if err != nil {
			log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), v)
			resultNode.Result = -1
			resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
			continue
		}

		if imdb.IsExistGroupMember(req.GroupID, v) {
			log.NewError(req.OperationID, "IsExistGroupMember ", req.GroupID, v)
			resultNode.Result = -1
			resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
			continue
		}
		var toInsertInfo db.GroupMember
		utils.CopyStructFields(&toInsertInfo, toUserInfo)
		toInsertInfo.GroupID = req.GroupID
		toInsertInfo.RoleLevel = constant.GroupOrdinaryUsers
		toInsertInfo.OperatorUserID = req.OpUserID
		err = imdb.InsertIntoGroupMember(toInsertInfo)
		if err != nil {
			log.NewError(req.OperationID, "InsertIntoGroupMember failed ", req.GroupID, toUserInfo.UserID, toUserInfo.Nickname, toUserInfo.FaceURL)
			resultNode.Result = -1
			resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
			continue
		}
		okUserIDList = append(okUserIDList, v)
		err = db.DB.AddGroupMember(req.GroupID, toUserInfo.UserID)
		if err != nil {
			log.NewError(req.OperationID, "AddGroupMember failed ", err.Error(), req.GroupID, toUserInfo.UserID)
		}
		resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
	}

	chat.MemberInvitedNotification(req.OperationID, req.GroupID, req.OpUserID, req.Reason, okUserIDList)
	resp.ErrCode = 0
	log.NewInfo(req.OperationID, "InviteUserToGroup rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) GetGroupAllMember(ctx context.Context, req *pbGroup.GetGroupAllMemberReq) (*pbGroup.GetGroupAllMemberResp, error) {
	log.NewInfo(req.OperationID, "GetGroupAllMember, args ", req.String())
	var resp pbGroup.GetGroupAllMemberResp
	memberList, err := imdb.GetGroupMemberListByGroupID(req.GroupID)
	if err != nil {
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		log.NewError(req.OperationID, "GetGroupMemberListByGroupID failed,", err.Error(), req.GroupID)
		return &resp, nil
	}

	for _, v := range memberList {
		var node open_im_sdk.GroupMemberFullInfo
		utils.CopyStructFields(&node, &v)
		resp.MemberList = append(resp.MemberList, &node)
	}
	log.NewInfo(req.OperationID, "GetGroupAllMember rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) GetGroupMemberList(ctx context.Context, req *pbGroup.GetGroupMemberListReq) (*pbGroup.GetGroupMemberListResp, error) {
	log.NewInfo(req.OperationID, "GetGroupMemberList args ", req.String())
	var resp pbGroup.GetGroupMemberListResp
	memberList, err := imdb.GetGroupMemberByGroupID(req.GroupID, req.Filter, req.NextSeq, 30)
	if err != nil {
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		log.NewError(req.OperationID, "GetGroupMemberByGroupId failed,", req.GroupID, req.Filter, req.NextSeq, 30)
		return &resp, nil
	}

	for _, v := range memberList {
		var node open_im_sdk.GroupMemberFullInfo
		utils.CopyStructFields(&node, &v)
		resp.MemberList = append(resp.MemberList, &node)
	}
	//db operate  get db sorted by join time
	if int32(len(memberList)) < 30 {
		resp.NextSeq = 0
	} else {
		resp.NextSeq = req.NextSeq + int32(len(memberList))
	}

	resp.ErrCode = 0
	log.NewInfo(req.OperationID, "GetGroupMemberList rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) KickGroupMember(ctx context.Context, req *pbGroup.KickGroupMemberReq) (*pbGroup.KickGroupMemberResp, error) {
	log.NewInfo(req.OperationID, "KickGroupMember args ", req.String())
	ownerList, err := imdb.GetOwnerManagerByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetOwnerManagerByGroupId failed ", err.Error(), req.GroupID)
		return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	//op is group owner?
	var flag = 0
	for _, v := range ownerList {
		if v.UserID == req.OpUserID {
			flag = 1
			log.NewDebug(req.OperationID, "is group owner ", req.OpUserID, req.GroupID)
			break
		}
	}

	//op is app manager
	if flag != 1 {
		if token_verify.IsMangerUserID(req.OpUserID) {
			flag = 1
			log.NewDebug(req.OperationID, "is app manager ", req.OpUserID)
		}
	}

	if flag != 1 {
		log.NewError(req.OperationID, "failed, no access kick ", req.OpUserID)
		return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}

	if len(req.KickedUserIDList) == 0 {
		log.NewError(req.OperationID, "failed, kick list 0")
		return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}, nil
	}

	groupOwnerUserID := ""
	for _, v := range ownerList {
		if v.RoleLevel == constant.GroupOwner {
			groupOwnerUserID = v.UserID
		}
	}

	var okUserIDList []string
	//remove
	var resp pbGroup.KickGroupMemberResp
	for _, v := range req.KickedUserIDList {
		//owner cant kicked
		if v == groupOwnerUserID {
			log.NewError(req.OperationID, "failed, can't kick owner ", v)
			resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
			continue
		}
		err := imdb.RemoveGroupMember(req.GroupID, v)
		if err != nil {
			log.NewError(req.OperationID, "RemoveGroupMember failed ", err.Error(), req.GroupID, v)
			resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
		} else {
			log.NewDebug(req.OperationID, "kicked ", v)
			resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: 0})
			okUserIDList = append(okUserIDList, v)
		}

		err = db.DB.DelGroupMember(req.GroupID, v)
		if err != nil {
			log.NewError(req.OperationID, "DelGroupMember failed ", err.Error(), req.GroupID, v)
		}
	}
	chat.MemberKickedNotification(req, okUserIDList)
	log.NewInfo(req.OperationID, "GetGroupMemberList rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) GetGroupMembersInfo(ctx context.Context, req *pbGroup.GetGroupMembersInfoReq) (*pbGroup.GetGroupMembersInfoResp, error) {
	log.NewInfo(req.OperationID, "GetGroupMembersInfo args ", req.String())

	var resp pbGroup.GetGroupMembersInfoResp

	for _, v := range req.MemberList {
		var memberNode open_im_sdk.GroupMemberFullInfo
		memberInfo, err := imdb.GetMemberInfoByID(req.GroupID, v)
		memberNode.UserID = v
		if err != nil {
			log.NewError(req.OperationID, "GetMemberInfoById failed ", err.Error(), req.GroupID, v)
			continue
		} else {
			utils.CopyStructFields(&memberNode, memberInfo)
			memberNode.JoinTime = memberInfo.JoinTime.Unix()
			resp.MemberList = append(resp.MemberList, &memberNode)
		}
	}
	resp.ErrCode = 0
	log.NewInfo(req.OperationID, "GetGroupMembersInfo rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) GetGroupApplicationList(_ context.Context, req *pbGroup.GetGroupApplicationListReq) (*pbGroup.GetGroupApplicationListResp, error) {
	log.NewInfo(req.OperationID, "GetGroupMembersInfo args ", req.String())
	reply, err := imdb.GetGroupApplicationList(req.FromUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupApplicationList failed ", err.Error(), req.FromUserID)
		return &pbGroup.GetGroupApplicationListResp{ErrCode: 701, ErrMsg: "GetGroupApplicationList failed"}, nil
	}

	log.NewDebug(req.OperationID, "GetGroupApplicationList reply ", reply)
	resp := pbGroup.GetGroupApplicationListResp{}
	for _, v := range reply {
		node := open_im_sdk.GroupRequest{UserInfo: &open_im_sdk.PublicUserInfo{}, GroupInfo: &open_im_sdk.GroupInfo{}}
		group, err := imdb.GetGroupInfoByGroupID(v.GroupID)
		if err != nil {
			log.Error(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), v.GroupID)
			continue
		}
		user, err := imdb.GetUserByUserID(v.UserID)
		if err != nil {
			log.Error(req.OperationID, "GetUserByUserID failed ", err.Error(), v.UserID)
			continue
		}

		cp.GroupRequestDBCopyOpenIM(&node, &v)
		cp.UserDBCopyOpenIMPublicUser(node.UserInfo, user)
		cp.GroupDBCopyOpenIM(node.GroupInfo, group)
		log.NewDebug(req.OperationID, "node ", node, "v ", v)
		resp.GroupRequestList = append(resp.GroupRequestList, &node)
	}
	log.NewInfo(req.OperationID, "GetGroupMembersInfo rpc return ", resp)
	return &resp, nil
}

func (s *groupServer) GetGroupsInfo(ctx context.Context, req *pbGroup.GetGroupsInfoReq) (*pbGroup.GetGroupsInfoResp, error) {
	log.NewInfo(req.OperationID, "GetGroupsInfo args ", req.String())
	groupsInfoList := make([]*open_im_sdk.GroupInfo, 0)
	for _, groupID := range req.GroupIDList {
		groupInfoFromMysql, err := imdb.GetGroupInfoByGroupID(groupID)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), groupID)
			continue
		}
		var groupInfo open_im_sdk.GroupInfo
		cp.GroupDBCopyOpenIM(&groupInfo, groupInfoFromMysql)
		groupsInfoList = append(groupsInfoList, &groupInfo)
	}

	resp := pbGroup.GetGroupsInfoResp{GroupInfoList: groupsInfoList}
	log.NewInfo(req.OperationID, "GetGroupsInfo rpc return ", resp)
	return &resp, nil
}

func (s *groupServer) GroupApplicationResponse(_ context.Context, req *pbGroup.GroupApplicationResponseReq) (*pbGroup.GroupApplicationResponseResp, error) {
	log.NewInfo(req.OperationID, "GroupApplicationResponse args ", req.String())

	groupRequest := db.GroupRequest{}
	utils.CopyStructFields(&groupRequest, req)
	groupRequest.UserID = req.FromUserID
	groupRequest.HandleUserID = req.OpUserID
	groupRequest.HandledTime = time.Now()
	if !token_verify.IsMangerUserID(req.OpUserID) && !imdb.IsGroupOwnerAdmin(req.GroupID, req.OpUserID) {
		log.NewError(req.OperationID, "IsMangerUserID IsGroupOwnerAdmin false ", req.GroupID, req.OpUserID)
		return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	err := imdb.UpdateGroupRequest(groupRequest)
	if err != nil {
		//{openIM002 7836e478bc43ce1d3b8889cac983f59b 1  ok 0001-01-01 00:00:00 +0000 UTC openIM001 0001-01-01 00:00:00 +0000 UTC }
		log.NewError(req.OperationID, "GroupApplicationResponse failed ", err.Error(), groupRequest)
		return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	if req.HandleResult == constant.GroupResponseAgree {
		user, err := imdb.GetUserByUserID(req.FromUserID)
		if err != nil {
			log.NewError(req.OperationID, "GroupApplicationResponse failed ", err.Error(), req.FromUserID)
			return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}
		member := db.GroupMember{}
		member.GroupID = req.GroupID
		member.UserID = req.FromUserID
		member.RoleLevel = constant.GroupOrdinaryUsers
		member.OperatorUserID = req.OpUserID
		member.FaceURL = user.FaceURL
		member.Nickname = user.Nickname

		err = imdb.InsertIntoGroupMember(member)
		if err != nil {
			log.NewError(req.OperationID, "GroupApplicationResponse failed ", err.Error(), member)
			return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}
		chat.GroupApplicationAcceptedNotification(req)
		chat.MemberEnterNotification(req)
	} else if req.HandleResult == constant.GroupResponseRefuse {
		chat.GroupApplicationRejectedNotification(req)
	} else {
		log.Error(req.OperationID, "HandleResult failed ", req.HandleResult)
		return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}

	log.NewInfo(req.OperationID, "rpc GroupApplicationResponse return ", pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{}})
	return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{}}, nil
}

func (s *groupServer) JoinGroup(ctx context.Context, req *pbGroup.JoinGroupReq) (*pbGroup.JoinGroupResp, error) {
	log.NewInfo(req.OperationID, "JoinGroup args ", req.String())
	_, err := imdb.GetUserByUserID(req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), req.OpUserID)
		return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	var groupRequest db.GroupRequest
	groupRequest.UserID = req.OpUserID
	groupRequest.ReqMsg = req.ReqMessage
	groupRequest.GroupID = req.GroupID

	err = imdb.InsertIntoGroupRequest(groupRequest)
	if err != nil {
		log.NewError(req.OperationID, "UpdateGroupRequest ", err.Error(), groupRequest)
		return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	_, err = imdb.GetGroupMemberListByGroupIDAndRoleLevel(req.GroupID, constant.GroupOwner)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberListByGroupIDAndRoleLevel failed ", err.Error(), req.GroupID, constant.GroupOwner)
		return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
	}

	chat.JoinGroupApplicationNotification(req)

	log.NewInfo(req.OperationID, "ReceiveJoinApplicationNotification rpc return ")
	return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

func (s *groupServer) QuitGroup(ctx context.Context, req *pbGroup.QuitGroupReq) (*pbGroup.QuitGroupResp, error) {
	log.NewError(req.OperationID, "QuitGroup args ", req.String())
	_, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberInfoByGroupIDAndUserID failed ", err.Error(), req.GroupID, req.OpUserID)
		return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	err = imdb.DeleteGroupMemberByGroupIDAndUserID(req.GroupID, req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "DeleteGroupMemberByGroupIdAndUserId failed ", err.Error(), req.GroupID, req.OpUserID)
		return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	err = db.DB.DelGroupMember(req.GroupID, req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "DelGroupMember failed ", req.GroupID, req.OpUserID)
		//	return &pbGroup.CommonResp{ErrorCode: constant.ErrQuitGroup.ErrCode, ErrorMsg: constant.ErrQuitGroup.ErrMsg}, nil
	}

	chat.MemberQuitNotification(req)
	log.NewInfo(req.OperationID, "rpc QuitGroup return ", pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}})
	return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

func hasAccess(req *pbGroup.SetGroupInfoReq) bool {
	if utils.IsContain(req.OpUserID, config.Config.Manager.AppManagerUid) {
		return true
	}
	groupUserInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupInfo.GroupID, req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberInfoByGroupIDAndUserID failed, ", err.Error(), req.GroupInfo.GroupID, req.OpUserID)
		return false

	}
	if groupUserInfo.RoleLevel == constant.GroupOwner || groupUserInfo.RoleLevel == constant.GroupAdmin {
		return true
	}
	return false
}

func (s *groupServer) SetGroupInfo(ctx context.Context, req *pbGroup.SetGroupInfoReq) (*pbGroup.SetGroupInfoResp, error) {
	log.NewInfo(req.OperationID, "SetGroupInfo args ", req.String())
	if !hasAccess(req) {
		log.NewError(req.OperationID, "no access ", req)
		return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}

	group, err := imdb.GetGroupInfoByGroupID(req.GroupInfo.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), req.GroupInfo.GroupID)
		return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, http.WrapError(constant.ErrDB)
	}

	////bitwise operators: 0001:groupName; 0010:Notification  0100:Introduction; 1000:FaceUrl; 10000:owner
	var changedType int32
	if group.GroupName != req.GroupInfo.GroupName && req.GroupInfo.GroupName != "" {
		changedType = 1
	}
	if group.Notification != req.GroupInfo.Notification && req.GroupInfo.Notification != "" {
		changedType = changedType | (1 << 1)
	}
	if group.Introduction != req.GroupInfo.Introduction && req.GroupInfo.Introduction != "" {
		changedType = changedType | (1 << 2)
	}
	if group.FaceURL != req.GroupInfo.FaceURL && req.GroupInfo.FaceURL != "" {
		changedType = changedType | (1 << 3)
	}
	//only administrators can set group information
	var groupInfo db.Group
	utils.CopyStructFields(&groupInfo, req.GroupInfo)
	err = imdb.SetGroupInfo(groupInfo)
	if err != nil {
		log.NewError(req.OperationID, "SetGroupInfo failed ", err.Error(), groupInfo)
		return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, http.WrapError(constant.ErrDB)
	}
	log.NewInfo(req.OperationID, "SetGroupInfo rpc return ", pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{}})
	if changedType != 0 {
		chat.GroupInfoSetNotification(req.OperationID, req.OpUserID, req.GroupInfo.GroupID)
	}
	return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{}}, nil
}

func (s *groupServer) TransferGroupOwner(_ context.Context, req *pbGroup.TransferGroupOwnerReq) (*pbGroup.TransferGroupOwnerResp, error) {
	log.NewInfo(req.OperationID, "TransferGroupOwner ", req.String())

	if req.OldOwnerUserID == req.NewOwnerUserID {
		log.NewError(req.OperationID, "same owner ", req.OldOwnerUserID, req.NewOwnerUserID)
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}
	groupMemberInfo := db.GroupMember{GroupID: req.GroupID, UserID: req.OldOwnerUserID, RoleLevel: constant.GroupOrdinaryUsers}
	err := imdb.UpdateGroupMemberInfo(groupMemberInfo)
	if err != nil {
		log.NewError(req.OperationID, "UpdateGroupMemberInfo failed ", groupMemberInfo)
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	groupMemberInfo = db.GroupMember{GroupID: req.GroupID, UserID: req.NewOwnerUserID, RoleLevel: constant.GroupOwner}
	err = imdb.UpdateGroupMemberInfo(groupMemberInfo)
	if err != nil {
		log.NewError(req.OperationID, "UpdateGroupMemberInfo failed ", groupMemberInfo)
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	chat.GroupOwnerTransferredNotification(req)

	return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil

}

func (s *groupServer) GetGroupById(_ context.Context, req *pbGroup.GetGroupByIdReq) (*pbGroup.GetGroupByIdResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbGroup.GetGroupByIdResp{CMSGroup: &pbGroup.CMSGroup{
		GroupInfo: &open_im_sdk.GroupInfo{},
	}}
	group, err := imdb.GetGroupById(req.GroupId)
	if err != nil {
		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "GetGroupById error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	resp.CMSGroup.GroupInfo = &open_im_sdk.GroupInfo{
		GroupID:       group.GroupID,
		GroupName:     group.GroupName,
		FaceURL:       group.FaceURL,
		OwnerUserID:   group.CreatorUserID,
		MemberCount:   0,
		Status:        group.Status,
		CreatorUserID: group.CreatorUserID,
		GroupType:     group.GroupType,
	}
	groupMember, err := imdb.GetGroupMaster(group.GroupID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMaster", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	resp.CMSGroup.GroupMasterName = groupMember.Nickname
	resp.CMSGroup.GroupMasterId = groupMember.UserID
	resp.CMSGroup.GroupInfo.CreatorUserID = group.CreatorUserID
	return resp, nil
}

func (s *groupServer) GetGroup(_ context.Context, req *pbGroup.GetGroupReq) (*pbGroup.GetGroupResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbGroup.GetGroupResp{
		CMSGroups: []*pbGroup.CMSGroup{},
	}
	groups, err := imdb.GetGroupsByName(req.GroupName, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupsByName error", req.String())
		return resp, http.WrapError(constant.ErrDB)
	}
	nums, err := imdb.GetGroupsCountNum(db.Group{GroupName: req.GroupName})
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupsCountNum error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	resp.GroupNums = nums
	resp.Pagination = &open_im_sdk.RequestPagination{
		PageNumber: req.Pagination.PageNumber,
		ShowNumber: req.Pagination.ShowNumber,
	}
	for _, v := range groups {
		groupMember, err := imdb.GetGroupMaster(v.GroupID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMaster error", err.Error())
		}
		resp.CMSGroups = append(resp.CMSGroups, &pbGroup.CMSGroup{
			GroupInfo: &open_im_sdk.GroupInfo{
				GroupID:       v.GroupID,
				GroupName:     v.GroupName,
				FaceURL:       v.FaceURL,
				OwnerUserID:   v.CreatorUserID,
				Status:        v.Status,
				CreatorUserID: v.CreatorUserID,
			},
			GroupMasterName: groupMember.Nickname,
			GroupMasterId:   groupMember.UserID,
		})
	}
	return resp, nil
}

func (s *groupServer) GetGroups(_ context.Context, req *pbGroup.GetGroupsReq) (*pbGroup.GetGroupsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "GetGroups ", req.String())
	resp := &pbGroup.GetGroupsResp{
		CMSGroups:  []*pbGroup.CMSGroup{},
		Pagination: &open_im_sdk.RequestPagination{},
	}
	groups, err := imdb.GetGroups(int(req.Pagination.PageNumber), int(req.Pagination.ShowNumber))
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroups error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	groupsCountNum, err := imdb.GetGroupsCountNum(db.Group{})
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "groupsCountNum ", groupsCountNum)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupsCountNum", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	resp.GroupNum = int32(groupsCountNum)
	resp.Pagination.PageNumber = req.Pagination.PageNumber
	resp.Pagination.ShowNumber = req.Pagination.ShowNumber
	for _, v := range groups {
		groupMember, err := imdb.GetGroupMaster(v.GroupID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		}
		resp.CMSGroups = append(resp.CMSGroups, &pbGroup.CMSGroup{
			GroupInfo: &open_im_sdk.GroupInfo{
				GroupID:       v.GroupID,
				GroupName:     v.GroupName,
				FaceURL:       v.FaceURL,
				OwnerUserID:   v.CreatorUserID,
				Status:        v.Status,
				CreatorUserID: v.CreatorUserID,
				CreateTime:    uint32(v.CreateTime.Unix()),
			},
			GroupMasterId:   groupMember.UserID,
			GroupMasterName: groupMember.Nickname,
		})
	}

	return resp, nil
}

func (s *groupServer) OperateGroupStatus(_ context.Context, req *pbGroup.OperateGroupStatusReq) (*pbGroup.OperateGroupStatusResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())
	resp := &pbGroup.OperateGroupStatusResp{}
	if err := imdb.OperateGroupStatus(req.GroupId, req.Status); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "OperateGroupStatus", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *groupServer) DeleteGroup(_ context.Context, req *pbGroup.DeleteGroupReq) (*pbGroup.DeleteGroupResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())
	resp := &pbGroup.DeleteGroupResp{}
	if err := imdb.DeleteGroup(req.GroupId); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteGroup error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *groupServer) OperateUserRole(_ context.Context, req *pbGroup.OperateUserRoleReq) (*pbGroup.OperateUserRoleResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args:", req.String())
	resp := &pbGroup.OperateUserRoleResp{}
	oldOwnerUserID, err := imdb.GetGroupMaster(req.GroupId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMaster failed", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName)
	client := pbGroup.NewGroupClient(etcdConn)
	var reqPb pbGroup.TransferGroupOwnerReq
	reqPb.OperationID = req.OperationID
	reqPb.NewOwnerUserID = req.UserId
	reqPb.GroupID = req.GroupId
	reqPb.OpUserID = "cms admin"
	reqPb.OldOwnerUserID = oldOwnerUserID.UserID
	reply, err := client.TransferGroupOwner(context.Background(), &reqPb)
	if reply.CommonResp.ErrCode != 0 || err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "TransferGroupOwner rpc failed")
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		}
	}
	return resp, nil
}

func (s *groupServer) GetGroupMembersCMS(_ context.Context, req *pbGroup.GetGroupMembersCMSReq) (*pbGroup.GetGroupMembersCMSResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args:", req.String())
	resp := &pbGroup.GetGroupMembersCMSResp{}
	groupMembers, err := imdb.GetGroupMembersByGroupIdCMS(req.GroupId, req.UserName, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMembersByGroupIdCMS Error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	groupMembersCount, err := imdb.GetGroupMembersCount(req.GroupId, req.UserName)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMembersCMS Error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	log.NewInfo(req.OperationID, groupMembersCount)
	resp.MemberNums = groupMembersCount
	for _, groupMember := range groupMembers {
		resp.Members = append(resp.Members, &open_im_sdk.GroupMemberFullInfo{
			GroupID:    req.GroupId,
			UserID:     groupMember.UserID,
			RoleLevel:  groupMember.RoleLevel,
			JoinTime:   groupMember.JoinTime.Unix(),
			Nickname:   groupMember.Nickname,
			FaceURL:    groupMember.FaceURL,
			JoinSource: groupMember.JoinSource,
		})
	}
	resp.Pagination = &open_im_sdk.ResponsePagination{
		CurrentPage: req.Pagination.PageNumber,
		ShowNumber:  req.Pagination.ShowNumber,
	}
	return resp, nil
}

func (s *groupServer) RemoveGroupMembersCMS(_ context.Context, req *pbGroup.RemoveGroupMembersCMSReq) (*pbGroup.RemoveGroupMembersCMSResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args:", req.String())
	resp := &pbGroup.RemoveGroupMembersCMSResp{}
	for _, userId := range req.UserIds {
		err := imdb.RemoveGroupMember(req.GroupId, userId)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
			resp.Failed = append(resp.Failed, userId)
		} else {
			resp.Success = append(resp.Success, userId)
		}
	}
	reqKick := &pbGroup.KickGroupMemberReq{
		GroupID:          req.GroupId,
		KickedUserIDList: resp.Success,
		Reason:           "admin kick",
		OperationID:      req.OperationID,
		OpUserID:         req.OpUserId,
	}
	chat.MemberKickedNotification(reqKick, resp.Success)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "success: ", resp.Success)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "failed: ", resp.Failed)
	return resp, nil
}

func (s *groupServer) AddGroupMembersCMS(_ context.Context, req *pbGroup.AddGroupMembersCMSReq) (*pbGroup.AddGroupMembersCMSResp, error) {
	log.NewInfo(req.OperationId, utils.GetSelfFuncName(), "args:", req.String())
	resp := &pbGroup.AddGroupMembersCMSResp{}
	for _, userId := range req.UserIds {
		if isExist := imdb.IsExistGroupMember(req.GroupId, userId); isExist {
			log.NewError(req.OperationId, utils.GetSelfFuncName(), "user is exist in group", userId, req.GroupId)
			resp.Failed = append(resp.Failed, userId)
			continue
		}
		user, err := imdb.GetUserByUserID(userId)
		if err != nil {
			log.NewError(req.OperationId, utils.GetSelfFuncName(), "GetUserByUserID", err.Error())
			resp.Failed = append(resp.Failed, userId)
			continue
		}
		groupMember := db.GroupMember{
			GroupID:        req.GroupId,
			UserID:         userId,
			Nickname:       user.Nickname,
			FaceURL:        "",
			RoleLevel:      1,
			JoinTime:       time.Time{},
			JoinSource:     constant.JoinByAdmin,
			OperatorUserID: "CmsAdmin",
			Ex:             "",
		}
		if err := imdb.InsertIntoGroupMember(groupMember); err != nil {
			log.NewError(req.OperationId, utils.GetSelfFuncName(), "InsertIntoGroupMember failed", req.String())
			resp.Failed = append(resp.Failed, userId)
		} else {
			resp.Success = append(resp.Success, userId)
		}
	}
	chat.MemberInvitedNotification(req.OperationId, req.GroupId, req.OpUserId, "admin add you to group", resp.Success)
	return resp, nil
}

func (s *groupServer) GetUserReqApplicationList(_ context.Context, req *pbGroup.GetUserReqApplicationListReq) (*pbGroup.GetUserReqApplicationListResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbGroup.GetUserReqApplicationListResp{}
	groupRequests, err := imdb.GetUserReqGroupByUserID(req.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserReqGroupByUserID failed ", err.Error())
		resp.CommonResp = &pbGroup.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return resp, nil
	}
	for _, groupReq := range groupRequests {
		node := open_im_sdk.GroupRequest{UserInfo: &open_im_sdk.PublicUserInfo{}, GroupInfo: &open_im_sdk.GroupInfo{}}
		group, err := imdb.GetGroupInfoByGroupID(groupReq.GroupID)
		if err != nil {
			log.Error(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), groupReq.GroupID)
			continue
		}
		user, err := imdb.GetUserByUserID(groupReq.UserID)
		if err != nil {
			log.Error(req.OperationID, "GetUserByUserID failed ", err.Error(), groupReq.UserID)
			continue
		}
		cp.GroupRequestDBCopyOpenIM(&node, &groupReq)
		cp.UserDBCopyOpenIMPublicUser(node.UserInfo, user)
		cp.GroupDBCopyOpenIM(node.GroupInfo, group)
		resp.GroupRequestList = append(resp.GroupRequestList, &node)
	}
	resp.CommonResp = &pbGroup.CommonResp{
		ErrCode: 0,
		ErrMsg:  "",
	}
	return resp, nil
}
