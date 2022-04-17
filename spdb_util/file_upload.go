package spdb_util

import (
	"bytes"
	"encoding/json"

	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ZZMarquis/gm/sm3"
	"github.com/ZZMarquis/gm/sm4"
	"github.com/ZZMarquis/gm/util"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

const(
	URL = "https://etest4.spdb.com.cn/spdb/uat/apiFile/upload" // 沙盒环境公共文件上传URL
)

// UploadFile 调用浦发公共文件上传接口
func UploadFile(security SpdbSecurity, fileReader io.Reader, fileSize int64, fileName string) (string, error) {
	//新建缓冲 用于存放文件内容
	bodyBuffer := &bytes.Buffer{}
	//创建multipart文件写入器，方便按照http规定格式写入内容
	bodyWriter := multipart.NewWriter(bodyBuffer)
	//从bodyWriter生成fileWriter,
	fileWriter, err := bodyWriter.CreateFormFile("s3File", fileName)
	if err != nil {
		return "", err
	}
	//将文件内容写入fileWriter
	_, err = io.Copy(fileWriter, fileReader)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	bodyWriter.Close()
	metaDataStr, _ := json.Marshal(FileUploadMetaData{
		FileName: fileName,
		FileSize: fmt.Sprintf("%dB", fileSize),
		FileSha1: Sha1(bodyBuffer.Bytes()),
	})
	header := map[string]string{
		"X-SPDB-Client-ID": security.ClientId,
		"X-SPDB-SM":        "true",
		"X-SPDB-LABEL":     "0001", //固定值
		"X-SPDB-SIGNATURE": sign(security.Secret, string(metaDataStr)),
		"X-SPDB-MetaData":  string(metaDataStr),
		"Content-Type":     bodyWriter.FormDataContentType(),
	}
	resp, err := doUpload(header, bodyBuffer)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func doUpload(header map[string]string, uploadBody io.Reader) (string, error) {
	client := http.DefaultClient
	req, err := http.NewRequest("POST", URL, uploadBody)
	if err != nil {
		logrus.Error("new request err, ", err)
		return "", err
	}
	// 设置header
	for k, v := range header {
		req.Header.Add(k, v)
	}
	//上传
	var respBody []byte
	response, err := client.Do(req)
	if err != nil {
		logrus.Error("client do err, ", err)
		return "", err
	}
	defer response.Body.Close()
	respBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.Error("read all err, ", err)
		return "", err
	}
	logrus.Info("http response: ", response)
	// http状态码200表示调用成功
	if response.StatusCode != 200 {
		logrus.Error("调用接口异常")
		return "", errors.New(string(respBody))
	}
	return string(respBody), err
}

func Sha1(data []byte) string {
	sha := sha1.New()
	sha.Write(data)
	return hex.EncodeToString(sha.Sum(nil))
}

// sign 对X-SPDB-MetaData加签
func sign(secret string, value string) string {
	sha := sha256.New()
	sha.Write([]byte(secret))
	sha256key := hex.EncodeToString(sha.Sum(nil))
	sm3Obj := sm3.New()
	sm3Obj.Write([]byte(sha256key))
	sm3Key := hex.EncodeToString(sm3Obj.Sum(nil))

	md5Key := md5.Sum([]byte(sm3Key))
	md5Key2 := md5Key[0:16]

	sha1Obj := sha1.New()
	sha1Obj.Write([]byte(value))
	dataByte := []byte(base64.StdEncoding.EncodeToString(sha1Obj.Sum(nil)))

	sm3Obj.Reset()
	sm3Obj.Write(dataByte)
	contentDigest := hex.EncodeToString(sm3Obj.Sum(nil))

	plainTextWithPadding := util.PKCS5Padding([]byte(contentDigest), sm4.BlockSize)
	cipherText, err := sm4.ECBEncrypt([]byte(md5Key2), plainTextWithPadding)
	if err != nil {
		logrus.Error("sm4 ECBEncrypt err, ", err)
	}
	encryptStr := base64.StdEncoding.EncodeToString([]byte((hex.EncodeToString(cipherText))))
	return encryptStr
}


