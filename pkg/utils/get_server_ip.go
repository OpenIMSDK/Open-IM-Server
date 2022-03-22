package utils

import (
	"Open_IM/pkg/common/config"
	"net"
)

// Deprecated: This value is no longer recommended.
// 不在建议使用该值：主要因为该值在每个组件部署时无法表示各自的实际ip，建议使用viper读取目标配置
//
// 比如：
//
// 需要读取rpc_auth地址时：viper.GetString("endpoints.rpc_auth")
var ServerIP = ""

func init() {
	//fixme In the configuration file, ip takes precedence, if not, get the valid network card ip of the machine
	if config.Config.ServerIP != "" {
		ServerIP = config.Config.ServerIP
		return
	}

	// see https://gist.github.com/jniltinho/9787946#gistcomment-3019898
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err.Error())
	}

	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ServerIP = localAddr.IP.String()
}
