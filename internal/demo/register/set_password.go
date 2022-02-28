package register

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	http2 "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type ParamsSetPassword struct {
	Email            string `json:"email"`
	PhoneNumber      string `json:"phoneNumber"`
	Password         string `json:"password"`
	VerificationCode string `json:"verificationCode"`
	Platform         int32  `json:"platform" binding:"required,min=1,max=7"`
	Ex               string `json:"ex"`
	OperationID      string `json:"operationID" binding:"required"`
}

func SetPassword(c *gin.Context) {
	params := ParamsSetPassword{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var account string
	if params.Email != "" {
		account = params.Email
	} else {
		account = params.PhoneNumber
	}
	if params.VerificationCode != config.Config.Demo.SuperCode {
		accountKey := account + "_" + constant.VerificationCodeForRegisterSuffix
		v, err := db.DB.GetAccountCode(accountKey)
		if err != nil || v != params.VerificationCode {
			log.NewError(params.OperationID, "password Verification code error", account, params.VerificationCode)
			data := make(map[string]interface{})
			data["PhoneNumber"] = account
			c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Verification code error!", "data": data})
			return
		}
	}
	url := fmt.Sprintf("http://%s:10000/auth/user_register", viper.GetString("endpoints.api"))
	openIMRegisterReq := api.UserRegisterReq{}
	openIMRegisterReq.OperationID = params.OperationID
	openIMRegisterReq.Platform = params.Platform
	openIMRegisterReq.UserID = account
	openIMRegisterReq.Nickname = account
	openIMRegisterReq.Secret = config.Config.Secret
	openIMRegisterResp := api.UserRegisterResp{}
	bMsg, err := http2.Post(url, openIMRegisterReq, config.Config.MessageCallBack.CallBackTimeOut)
	if err != nil {
		log.NewError(params.OperationID, "request openIM register error", account, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": err.Error()})
		return
	}
	err = json.Unmarshal(bMsg, &openIMRegisterResp)
	if err != nil || openIMRegisterResp.ErrCode != 0 {
		log.NewError(params.OperationID, "request openIM register error", account, "err", "resp: ", openIMRegisterResp.ErrCode)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), err.Error())
		}
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": "register failed: " + openIMRegisterResp.ErrMsg})
		return
	}
	log.Info(params.OperationID, "begin store mysql", account, params.Password)
	err = im_mysql_model.SetPassword(account, params.Password, params.Ex)
	if err != nil {
		log.NewError(params.OperationID, "set phone number password error", account, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": err.Error()})
		return
	}
	log.Info(params.OperationID, "end setPassword", account, params.Password)
	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": openIMRegisterResp.UserToken})
	return
}
