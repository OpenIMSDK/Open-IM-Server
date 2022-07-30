package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

var (
	_, b, _, _ = runtime.Caller(0)
	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b), "../../..")
)

var Config config

type callBackConfig struct {
	Enable                 bool `yaml:"enable"`
	CallbackTimeOut        int  `yaml:"callbackTimeOut"`
	CallbackFailedContinue bool `yaml:"callbackFailedContinue"`
}

type config struct {
	ServerIP string `yaml:"serverip"`

	RpcRegisterIP string `yaml:"rpcRegisterIP"`
	ListenIP      string `yaml:"listenIP"`

	ServerVersion string `yaml:"serverversion"`
	Api           struct {
		GinPort  []int  `yaml:"openImApiPort"`
		ListenIP string `yaml:"listenIP"`
	}
	CmsApi struct {
		GinPort  []int  `yaml:"openImCmsApiPort"`
		ListenIP string `yaml:"listenIP"`
	}
	Sdk struct {
		WsPort  []int    `yaml:"openImSdkWsPort"`
		DataDir []string `yaml:"dataDir"`
	}
	Credential struct {
		Tencent struct {
			AppID     string `yaml:"appID"`
			Region    string `yaml:"region"`
			Bucket    string `yaml:"bucket"`
			SecretID  string `yaml:"secretID"`
			SecretKey string `yaml:"secretKey"`
		}
		Ali struct {
			RegionID           string `yaml:"regionID"`
			AccessKeyID        string `yaml:"accessKeyID"`
			AccessKeySecret    string `yaml:"accessKeySecret"`
			StsEndpoint        string `yaml:"stsEndpoint"`
			OssEndpoint        string `yaml:"ossEndpoint"`
			Bucket             string `yaml:"bucket"`
			FinalHost          string `yaml:"finalHost"`
			StsDurationSeconds int64  `yaml:"stsDurationSeconds"`
			OssRoleArn         string `yaml:"OssRoleArn"`
		}
		Minio struct {
			Bucket              string `yaml:"bucket"`
			AppBucket           string `yaml:"appBucket"`
			Location            string `yaml:"location"`
			Endpoint            string `yaml:"endpoint"`
			AccessKeyID         string `yaml:"accessKeyID"`
			SecretAccessKey     string `yaml:"secretAccessKey"`
			EndpointInner       string `yaml:"endpointInner"`
			EndpointInnerEnable bool   `yaml:"endpointInnerEnable"`
			StorageTime         int    `yaml:"storageTime"`
		} `yaml:"minio"`
		Aws struct {
			AccessKeyID     string `yaml:"accessKeyID"`
			AccessKeySecret string `yaml:"accessKeySecret"`
			Region          string `yaml:"region"`
			Bucket          string `yaml:"bucket"`
			FinalHost       string `yaml:"finalHost"`
			RoleArn         string `yaml:"roleArn"`
			ExternalId      string `yaml:"externalId"`
			RoleSessionName string `yaml:"roleSessionName"`
		} `yaml:"aws"`
	}

	Dtm struct {
		ServerURL string `json:"serverURL"`
	}

	Mysql struct {
		DBAddress      []string `yaml:"dbMysqlAddress"`
		DBUserName     string   `yaml:"dbMysqlUserName"`
		DBPassword     string   `yaml:"dbMysqlPassword"`
		DBDatabaseName string   `yaml:"dbMysqlDatabaseName"`
		DBTableName    string   `yaml:"DBTableName"`
		DBMsgTableNum  int      `yaml:"dbMsgTableNum"`
		DBMaxOpenConns int      `yaml:"dbMaxOpenConns"`
		DBMaxIdleConns int      `yaml:"dbMaxIdleConns"`
		DBMaxLifeTime  int      `yaml:"dbMaxLifeTime"`
	}
	Mongo struct {
		DBUri               string `yaml:"dbUri"`
		DBAddress           string `yaml:"dbAddress"`
		DBDirect            bool   `yaml:"dbDirect"`
		DBTimeout           int    `yaml:"dbTimeout"`
		DBDatabase          string `yaml:"dbDatabase"`
		DBSource            string `yaml:"dbSource"`
		DBUserName          string `yaml:"dbUserName"`
		DBPassword          string `yaml:"dbPassword"`
		DBMaxPoolSize       int    `yaml:"dbMaxPoolSize"`
		DBRetainChatRecords int    `yaml:"dbRetainChatRecords"`
	}
	Redis struct {
		DBAddress     []string `yaml:"dbAddress"`
		DBMaxIdle     int      `yaml:"dbMaxIdle"`
		DBMaxActive   int      `yaml:"dbMaxActive"`
		DBIdleTimeout int      `yaml:"dbIdleTimeout"`
		DBUserName    string   `yaml:"dbUserName"`
		DBPassWord    string   `yaml:"dbPassWord"`
		EnableCluster bool     `yaml:"enableCluster"`
	}
	RpcPort struct {
		OpenImUserPort           []int `yaml:"openImUserPort"`
		OpenImFriendPort         []int `yaml:"openImFriendPort"`
		OpenImMessagePort        []int `yaml:"openImMessagePort"`
		OpenImMessageGatewayPort []int `yaml:"openImMessageGatewayPort"`
		OpenImGroupPort          []int `yaml:"openImGroupPort"`
		OpenImAuthPort           []int `yaml:"openImAuthPort"`
		OpenImPushPort           []int `yaml:"openImPushPort"`
		OpenImStatisticsPort     []int `yaml:"openImStatisticsPort"`
		OpenImMessageCmsPort     []int `yaml:"openImMessageCmsPort"`
		OpenImAdminCmsPort       []int `yaml:"openImAdminCmsPort"`
		OpenImOfficePort         []int `yaml:"openImOfficePort"`
		OpenImOrganizationPort   []int `yaml:"openImOrganizationPort"`
		OpenImConversationPort   []int `yaml:"openImConversationPort"`
		OpenImCachePort          []int `yaml:"openImCachePort"`
	}
	RpcRegisterName struct {
		OpenImStatisticsName string `yaml:"openImStatisticsName"`
		OpenImUserName       string `yaml:"openImUserName"`
		OpenImFriendName     string `yaml:"openImFriendName"`
		//	OpenImOfflineMessageName     string `yaml:"openImOfflineMessageName"`
		OpenImMsgName          string `yaml:"openImMsgName"`
		OpenImPushName         string `yaml:"openImPushName"`
		OpenImRelayName        string `yaml:"openImRelayName"`
		OpenImGroupName        string `yaml:"openImGroupName"`
		OpenImAuthName         string `yaml:"openImAuthName"`
		OpenImMessageCMSName   string `yaml:"openImMessageCMSName"`
		OpenImAdminCMSName     string `yaml:"openImAdminCMSName"`
		OpenImOfficeName       string `yaml:"openImOfficeName"`
		OpenImOrganizationName string `yaml:"openImOrganizationName"`
		OpenImConversationName string `yaml:"openImConversationName"`
		OpenImCacheName        string `yaml:"openImCacheName"`
		OpenImRealTimeCommName string `yaml:"openImRealTimeCommName"`
	}
	Etcd struct {
		EtcdSchema string   `yaml:"etcdSchema"`
		EtcdAddr   []string `yaml:"etcdAddr"`
	}
	Log struct {
		StorageLocation       string   `yaml:"storageLocation"`
		RotationTime          int      `yaml:"rotationTime"`
		RemainRotationCount   uint     `yaml:"remainRotationCount"`
		RemainLogLevel        uint     `yaml:"remainLogLevel"`
		ElasticSearchSwitch   bool     `yaml:"elasticSearchSwitch"`
		ElasticSearchAddr     []string `yaml:"elasticSearchAddr"`
		ElasticSearchUser     string   `yaml:"elasticSearchUser"`
		ElasticSearchPassword string   `yaml:"elasticSearchPassword"`
	}
	ModuleName struct {
		LongConnSvrName string `yaml:"longConnSvrName"`
		MsgTransferName string `yaml:"msgTransferName"`
		PushName        string `yaml:"pushName"`
	}
	LongConnSvr struct {
		WebsocketPort       []int `yaml:"openImWsPort"`
		WebsocketMaxConnNum int   `yaml:"websocketMaxConnNum"`
		WebsocketMaxMsgLen  int   `yaml:"websocketMaxMsgLen"`
		WebsocketTimeOut    int   `yaml:"websocketTimeOut"`
	}

	Push struct {
		Tpns struct {
			Ios struct {
				AccessID  string `yaml:"accessID"`
				SecretKey string `yaml:"secretKey"`
			}
			Android struct {
				AccessID  string `yaml:"accessID"`
				SecretKey string `yaml:"secretKey"`
			}
			Enable bool `yaml:"enable"`
		}
		Jpns struct {
			AppKey       string `yaml:"appKey"`
			MasterSecret string `yaml:"masterSecret"`
			PushUrl      string `yaml:"pushUrl"`
			PushIntent   string `yaml:"pushIntent"`
			Enable       bool   `yaml:"enable"`
		}
		Getui struct {
			PushUrl      string `yaml:"pushUrl"`
			AppKey       string `yaml:"appKey"`
			Enable       bool   `yaml:"enable"`
			Intent       string `yaml:"intent"`
			MasterSecret string `yaml:"masterSecret"`
		}
		Fcm struct {
			ServiceAccount string `yaml:"serviceAccount"`
			Enable         bool   `yaml:"enable"`
		}
	}
	Manager struct {
		AppManagerUid          []string `yaml:"appManagerUid"`
		Secrets                []string `yaml:"secrets"`
		AppSysNotificationName string   `yaml:"appSysNotificationName"`
	}

	Kafka struct {
		Ws2mschat struct {
			Addr  []string `yaml:"addr"`
			Topic string   `yaml:"topic"`
		}
		Ws2mschatOffline struct {
			Addr  []string `yaml:"addr"`
			Topic string   `yaml:"topic"`
		}
		MsgToMongo struct {
			Addr  []string `yaml:"addr"`
			Topic string   `yaml:"topic"`
		}
		Ms2pschat struct {
			Addr  []string `yaml:"addr"`
			Topic string   `yaml:"topic"`
		}
		ConsumerGroupID struct {
			MsgToRedis string `yaml:"msgToTransfer"`
			MsgToMongo string `yaml:"msgToMongo"`
			MsgToMySql string `yaml:"msgToMySql"`
			MsgToPush  string `yaml:"msgToPush"`
		}
	}
	Secret                            string `yaml:"secret"`
	MultiLoginPolicy                  int    `yaml:"multiloginpolicy"`
	ChatPersistenceMysql              bool   `yaml:"chatpersistencemysql"`
	ReliableStorage                   bool   `yaml:"reliablestorage"`
	MsgCacheTimeout                   int    `yaml:"msgCacheTimeout"`
	GroupMessageHasReadReceiptEnable  bool   `yaml:"groupMessageHasReadReceiptEnable"`
	SingleMessageHasReadReceiptEnable bool   `yaml:"singleMessageHasReadReceiptEnable"`

	TokenPolicy struct {
		AccessSecret string `yaml:"accessSecret"`
		AccessExpire int64  `yaml:"accessExpire"`
	}
	MessageVerify struct {
		FriendVerify bool `yaml:"friendVerify"`
	}
	IOSPush struct {
		PushSound  string `yaml:"pushSound"`
		BadgeCount bool   `yaml:"badgeCount"`
		Production bool   `yaml:"production"`
	}

	Callback struct {
		CallbackUrl                        string         `yaml:"callbackUrl"`
		CallbackBeforeSendSingleMsg        callBackConfig `yaml:"callbackBeforeSendSingleMsg"`
		CallbackAfterSendSingleMsg         callBackConfig `yaml:"callbackAfterSendSingleMsg"`
		CallbackBeforeSendGroupMsg         callBackConfig `yaml:"callbackBeforeSendGroupMsg"`
		CallbackAfterSendGroupMsg          callBackConfig `yaml:"callbackAfterSendGroupMsg"`
		CallbackWordFilter                 callBackConfig `yaml:"callbackWordFilter"`
		CallbackUserOnline                 callBackConfig `yaml:"callbackUserOnline"`
		CallbackUserOffline                callBackConfig `yaml:"callbackUserOffline"`
		CallbackOfflinePush                callBackConfig `yaml:"callbackOfflinePush"`
		CallbackOnlinePush                 callBackConfig `yaml:"callbackOnlinePush"`
		CallbackBeforeSuperGroupOnlinePush callBackConfig `yaml:"callbackSuperGroupOnlinePush"`
	} `yaml:"callback"`
	Notification struct {
		///////////////////////group/////////////////////////////
		GroupCreated struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupCreated"`

		GroupInfoSet struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupInfoSet"`

		JoinGroupApplication struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"joinGroupApplication"`

		MemberQuit struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"memberQuit"`

		GroupApplicationAccepted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupApplicationAccepted"`

		GroupApplicationRejected struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupApplicationRejected"`

		GroupOwnerTransferred struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupOwnerTransferred"`

		MemberKicked struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"memberKicked"`

		MemberInvited struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"memberInvited"`

		MemberEnter struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"memberEnter"`

		GroupDismissed struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupDismissed"`

		GroupMuted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMuted"`

		GroupCancelMuted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupCancelMuted"`

		GroupMemberMuted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberMuted"`

		GroupMemberCancelMuted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberCancelMuted"`
		GroupMemberInfoSet struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberInfoSet"`
		GroupMemberSetToAdmin struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberSetToAdmin"`
		GroupMemberSetToOrdinary struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberSetToOrdinaryUser"`
		OrganizationChanged struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"organizationChanged"`

		////////////////////////user///////////////////////
		UserInfoUpdated struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"userInfoUpdated"`

		//////////////////////friend///////////////////////
		FriendApplication struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendApplicationAdded"`
		FriendApplicationApproved struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendApplicationApproved"`

		FriendApplicationRejected struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendApplicationRejected"`

		FriendAdded struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendAdded"`

		FriendDeleted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendDeleted"`
		FriendRemarkSet struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendRemarkSet"`
		BlackAdded struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"blackAdded"`
		BlackDeleted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"blackDeleted"`
		ConversationOptUpdate struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"conversationOptUpdate"`
		ConversationSetPrivate struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  struct {
				OpenTips  string `yaml:"openTips"`
				CloseTips string `yaml:"closeTips"`
			} `yaml:"defaultTips"`
		} `yaml:"conversationSetPrivate"`
		WorkMomentsNotification struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"workMomentsNotification"`
		JoinDepartmentNotification struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"joinDepartmentNotification"`
		Signal struct {
			OfflinePush struct {
				Title string `yaml:"title"`
			} `yaml:"offlinePush"`
		} `yaml:"signal"`
	}
	Demo struct {
		Port         []int  `yaml:"openImDemoPort"`
		ListenIP     string `yaml:"listenIP"`
		AliSMSVerify struct {
			AccessKeyID                  string `yaml:"accessKeyId"`
			AccessKeySecret              string `yaml:"accessKeySecret"`
			SignName                     string `yaml:"signName"`
			VerificationCodeTemplateCode string `yaml:"verificationCodeTemplateCode"`
			Enable                       bool   `yaml:"enable"`
		}
		TencentSMS struct {
			AppID                        string `yaml:"appID"`
			Region                       string `yaml:"region"`
			SecretID                     string `yaml:"secretID"`
			SecretKey                    string `yaml:"secretKey"`
			SignName                     string `yaml:"signName"`
			VerificationCodeTemplateCode string `yaml:"verificationCodeTemplateCode"`
			Enable                       bool   `yaml:"enable"`
		}
		SuperCode    string `yaml:"superCode"`
		CodeTTL      int    `yaml:"codeTTL"`
		UseSuperCode bool   `yaml:"useSuperCode"`
		Mail         struct {
			Title                   string `yaml:"title"`
			SenderMail              string `yaml:"senderMail"`
			SenderAuthorizationCode string `yaml:"senderAuthorizationCode"`
			SmtpAddr                string `yaml:"smtpAddr"`
			SmtpPort                int    `yaml:"smtpPort"`
		}
		TestDepartMentID string `yaml:"testDepartMentID"`
		ImAPIURL         string `yaml:"imAPIURL"`
	}
	Rtc struct {
		SignalTimeout string `yaml:"signalTimeout"`
	} `yaml:"rtc"`
}
type PConversation struct {
	ReliabilityLevel int  `yaml:"reliabilityLevel"`
	UnreadCount      bool `yaml:"unreadCount"`
}

type POfflinePush struct {
	PushSwitch bool   `yaml:"switch"`
	Title      string `yaml:"title"`
	Desc       string `yaml:"desc"`
	Ext        string `yaml:"ext"`
}
type PDefaultTips struct {
	Tips string `yaml:"tips"`
}

func init() {
	cfgName := os.Getenv("CONFIG_NAME")
	fmt.Println(Root, cfgName)

	if len(cfgName) == 0 {
		cfgName = Root + "/config/config.yaml"
	}

	bytes, err := ioutil.ReadFile(cfgName)
	if err != nil {
		panic(err.Error())
	}
	if err = yaml.Unmarshal(bytes, &Config); err != nil {
		panic(err.Error())
	}
}
