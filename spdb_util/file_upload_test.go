package spdb_util

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"testing"
)

func TestUploadFile(t *testing.T) {
	fileData, err := ioutil.ReadFile("test.txt")
	if err != nil {
		t.Error(err)
		return
	}
	security := SpdbSecurity{
		ClientId: "****",
		Secret:   "****",
	}
	resp, err := UploadFile(security, bytes.NewBuffer(fileData), int64(len(fileData)), "test.txt")
	if err != nil {
		t.Error(err)
		return
	}
	logrus.Info("resp: ", resp)
}