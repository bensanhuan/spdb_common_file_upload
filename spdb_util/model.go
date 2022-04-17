package spdb_util

// SpdbSecurity 浦发API接口相关参数
type SpdbSecurity struct {
	// ClientId 浦发API平台APP唯一标识X-SPDB-Client-ID
	ClientId string
	// Secret 浦发API平台secret，用于加密
	Secret string

	// 一下公私钥，全报文加密才需要
	//// 合作方sm2私钥，用于加签
	//PrivateKey string
	//// 浦发sm2公钥，用于验签
	//SpdbPublicKey string
}

// 公共文件上传
// FileUploadMetaData 元数据
type FileUploadMetaData struct {
	FileName string `json:"fileName"`
	FileSize string `json:"fileSize"`
	FileSha1 string `json:"fileSha1"`
}
