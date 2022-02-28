package apiThird

import (
	apiStruct "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	_ "Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/minio/minio-go/v7"
	cr "github.com/minio/minio-go/v7/pkg/credentials"
)

func MinioStorageCredential(c *gin.Context) {
	var (
		req  apiStruct.MinioStorageCredentialReq
		resp apiStruct.MiniostorageCredentialResp
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	//ok, _ := token_verify.GetUserIDFromToken(c.Request.Header.Get("token"))
	//if !ok {
	//	log.NewError("", utils.GetSelfFuncName(), "GetUserIDFromToken false ", c.Request.Header.Get("token"))
	//	c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": "GetUserIDFromToken failed"})
	//	return
	//}
	var stsOpts cr.STSAssumeRoleOptions
	stsOpts.AccessKey = config.Config.Credential.Minio.AccessKeyID
	stsOpts.SecretKey = config.Config.Credential.Minio.SecretAccessKey
	stsOpts.DurationSeconds = constant.MinioDurationTimes
	li, err := cr.NewSTSAssumeRole(config.Config.Credential.Minio.Endpoint, stsOpts)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "NewSTSAssumeRole failed", err.Error(), stsOpts, config.Config.Credential.Minio.Endpoint)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	v, err := li.Get()
	if err != nil {
		log.NewError("0", utils.GetSelfFuncName(), "li.Get error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	if err != nil {
		log.NewError("0", utils.GetSelfFuncName(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	resp.SessionToken = v.SessionToken
	resp.SecretAccessKey = v.SecretAccessKey
	resp.AccessKeyID = v.AccessKeyID
	resp.BucketName = config.Config.Credential.Minio.Bucket
	resp.StsEndpointURL = config.Config.Credential.Minio.Endpoint
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": resp})
}
