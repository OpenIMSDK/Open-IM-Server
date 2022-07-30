package base_info

type AwsStorageCredentialReq struct {
	OperationID string `json:"operationID"`
}

type AwsStorageCredentialRespData struct {
	AccessKeyId     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	RegionID        string `json:"regionId"`
	Bucket          string `json:"bucket"`
	FinalHost       string `json:"FinalHost"`
}

type AwsStorageCredentialResp struct {
	CommResp
	CosData AwsStorageCredentialRespData
	Data    map[string]interface{} `json:"data"`
}
