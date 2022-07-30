package constant

const (

	//group admin
	//	OrdinaryMember = 0
	//	GroupOwner     = 1
	//	Administrator  = 2
	//group application
	//	Application      = 0
	//	AgreeApplication = 1

	//friend related
	BlackListFlag         = 1
	ApplicationFriendFlag = 0
	FriendFlag            = 1
	RefuseFriendFlag      = -1

	//Websocket Protocol
	WSGetNewestSeq     = 1001
	WSPullMsgBySeqList = 1002
	WSSendMsg          = 1003
	WSSendSignalMsg    = 1004
	WSPushMsg          = 2001
	WSKickOnlineMsg    = 2002
	WsLogoutMsg        = 2003
	WSDataError        = 3001

	///ContentType
	//UserRelated
	Text                         = 101
	Picture                      = 102
	Voice                        = 103
	Video                        = 104
	File                         = 105
	AtText                       = 106
	Merger                       = 107
	Card                         = 108
	Location                     = 109
	Custom                       = 110
	Revoke                       = 111
	HasReadReceipt               = 112
	Typing                       = 113
	Quote                        = 114
	GroupHasReadReceipt          = 116
	AdvancedText                 = 117
	AdvancedRevoke               = 118 //影响前者消息
	CustomNotTriggerConversation = 119
	CustomOnlineOnly             = 120

	Common             = 200
	GroupMsg           = 201
	SignalMsg          = 202
	CustomNotification = 203

	//SysRelated
	NotificationBegin                     = 1000
	DeleteMessageNotification             = 1100
	FriendApplicationApprovedNotification = 1201 //add_friend_response
	FriendApplicationRejectedNotification = 1202 //add_friend_response
	FriendApplicationNotification         = 1203 //add_friend
	FriendAddedNotification               = 1204
	FriendDeletedNotification             = 1205 //delete_friend
	FriendRemarkSetNotification           = 1206 //set_friend_remark?
	BlackAddedNotification                = 1207 //add_black
	BlackDeletedNotification              = 1208 //remove_black
	MsgDeleteNotification                 = 1209

	ConversationOptChangeNotification = 1300 // change conversation opt

	UserNotificationBegin       = 1301
	UserInfoUpdatedNotification = 1303 //SetSelfInfoTip             = 204
	UserNotificationEnd         = 1399
	OANotification              = 1400

	GroupNotificationBegin = 1500

	GroupCreatedNotification                 = 1501
	GroupInfoSetNotification                 = 1502
	JoinGroupApplicationNotification         = 1503
	MemberQuitNotification                   = 1504
	GroupApplicationAcceptedNotification     = 1505
	GroupApplicationRejectedNotification     = 1506
	GroupOwnerTransferredNotification        = 1507
	MemberKickedNotification                 = 1508
	MemberInvitedNotification                = 1509
	MemberEnterNotification                  = 1510
	GroupDismissedNotification               = 1511
	GroupMemberMutedNotification             = 1512
	GroupMemberCancelMutedNotification       = 1513
	GroupMutedNotification                   = 1514
	GroupCancelMutedNotification             = 1515
	GroupMemberInfoSetNotification           = 1516
	GroupMemberSetToAdminNotification        = 1517
	GroupMemberSetToOrdinaryUserNotification = 1518

	SignalingNotificationBegin = 1600
	SignalingNotification      = 1601
	SignalingNotificationEnd   = 1649

	SuperGroupNotificationBegin  = 1650
	SuperGroupUpdateNotification = 1651
	SuperGroupNotificationEnd    = 1699

	ConversationPrivateChatNotification = 1701

	OrganizationChangedNotification = 1801

	WorkMomentNotificationBegin = 1900
	WorkMomentNotification      = 1901

	NotificationEnd = 3000

	//status
	MsgNormal  = 1
	MsgDeleted = 4

	//MsgFrom
	UserMsgType = 100
	SysMsgType  = 200

	//SessionType
	SingleChatType       = 1
	GroupChatType        = 2
	SuperGroupChatType   = 3
	NotificationChatType = 4
	//token
	NormalToken  = 0
	InValidToken = 1
	KickedToken  = 2
	ExpiredToken = 3

	//MultiTerminalLogin
	//Full-end login, but the same end is mutually exclusive
	AllLoginButSameTermKick = 1
	//Only one of the endpoints can log in
	SingleTerminalLogin = 2
	//The web side can be online at the same time, and the other side can only log in at one end
	WebAndOther = 3
	//The PC side is mutually exclusive, and the mobile side is mutually exclusive, but the web side can be online at the same time
	PcMobileAndWeb = 4

	OnlineStatus  = "online"
	OfflineStatus = "offline"
	Registered    = "registered"
	UnRegistered  = "unregistered"

	//MsgReceiveOpt
	ReceiveMessage          = 0
	NotReceiveMessage       = 1
	ReceiveNotNotifyMessage = 2

	//OptionsKey
	IsHistory                  = "history"
	IsPersistent               = "persistent"
	IsOfflinePush              = "offlinePush"
	IsUnreadCount              = "unreadCount"
	IsConversationUpdate       = "conversationUpdate"
	IsSenderSync               = "senderSync"
	IsNotPrivate               = "notPrivate"
	IsSenderConversationUpdate = "senderConversationUpdate"
	IsSenderNotificationPush   = "senderNotificationPush"

	//GroupStatus
	GroupOk              = 0
	GroupBanChat         = 1
	GroupStatusDismissed = 2
	GroupStatusMuted     = 3

	//GroupType
	NormalGroup  = 0
	SuperGroup   = 1
	WorkingGroup = 2

	GroupBaned          = 3
	GroupBanPrivateChat = 4

	//UserJoinGroupSource
	JoinByAdmin = 1

	JoinByInvitation = 2
	JoinBySearch     = 3
	JoinByQRCode     = 4

	//Minio
	MinioDurationTimes = 3600
	//Aws
	AwsDurationTimes = 3600
	// verificationCode used for
	VerificationCodeForRegister       = 1
	VerificationCodeForReset          = 2
	VerificationCodeForRegisterSuffix = "_forRegister"
	VerificationCodeForResetSuffix    = "_forReset"

	//callbackCommand
	CallbackBeforeSendSingleMsgCommand  = "callbackBeforeSendSingleMsgCommand"
	CallbackAfterSendSingleMsgCommand   = "callbackAfterSendSingleMsgCommand"
	CallbackBeforeSendGroupMsgCommand   = "callbackBeforeSendGroupMsgCommand"
	CallbackAfterSendGroupMsgCommand    = "callbackAfterSendGroupMsgCommand"
	CallbackWordFilterCommand           = "callbackWordFilterCommand"
	CallbackUserOnlineCommand           = "callbackUserOnlineCommand"
	CallbackUserOfflineCommand          = "callbackUserOfflineCommand"
	CallbackOfflinePushCommand          = "callbackOfflinePushCommand"
	CallbackOnlinePushCommand           = "callbackOnlinePushCommand"
	CallbackSuperGroupOnlinePushCommand = "callbackSuperGroupOnlinePushCommand"
	//callback actionCode
	ActionAllow     = 0
	ActionForbidden = 1
	//callback callbackHandleCode
	CallbackHandleSuccess = 0
	CallbackHandleFailed  = 1

	// minioUpload
	OtherType = 1
	VideoType = 2
	ImageType = 3

	// workMoment permission
	WorkMomentPublic            = 0
	WorkMomentPrivate           = 1
	WorkMomentPermissionCanSee  = 2
	WorkMomentPermissionCantSee = 3

	// workMoment sdk notification type
	WorkMomentCommentNotification = 0
	WorkMomentLikeNotification    = 1
	WorkMomentAtUserNotification  = 2
)
const (
	AtAllString       = "AtAllTag"
	AtNormal          = 0
	AtMe              = 1
	AtAll             = 2
	AtAllAtMe         = 3
	GroupNotification = 4
)

var ContentType2PushContent = map[int64]string{
	Picture:   "[图片]",
	Voice:     "[语音]",
	Video:     "[视频]",
	File:      "[文件]",
	Text:      "你收到了一条文本消息",
	AtText:    "[有人@你]",
	GroupMsg:  "你收到一条群聊消息",
	Common:    "你收到一条新消息",
	SignalMsg: "音视频通话邀请",
}

const (
	FieldRecvMsgOpt    = 1
	FieldIsPinned      = 2
	FieldAttachedInfo  = 3
	FieldIsPrivateChat = 4
	FieldGroupAtType   = 5
	FieldIsNotInGroup  = 6
	FieldEx            = 7
	FieldUnread        = 8
)

const (
	AppOrdinaryUsers = 1
	AppAdmin         = 2

	GroupOrdinaryUsers = 1
	GroupOwner         = 2
	GroupAdmin         = 3

	GroupResponseAgree  = 1
	GroupResponseRefuse = -1

	FriendResponseAgree  = 1
	FriendResponseRefuse = -1

	Male   = 1
	Female = 2
)

const (
	UnreliableNotification    = 1
	ReliableNotificationNoMsg = 2
	ReliableNotificationMsg   = 3
)

const (
	ApplyNeedVerificationInviteDirectly = 0 // 申请需要同意 邀请直接进
	AllNeedVerification                 = 1 //所有人进群需要验证，除了群主管理员邀请进群
	Directly                            = 2 //直接进群
)

const FriendAcceptTip = "You have successfully become friends, so start chatting"

func GroupIsBanChat(status int32) bool {
	if status != GroupStatusMuted {
		return false
	}
	return true
}

func GroupIsBanPrivateChat(status int32) bool {
	if status != GroupBanPrivateChat {
		return false
	}
	return true
}

const (
	TokenKicked = 1001
)

const BigVersion = "v2"

const LogFileName = "OpenIM.log"

const StatisticsTimeInterval = 60

const MaxNotificationNum = 100
