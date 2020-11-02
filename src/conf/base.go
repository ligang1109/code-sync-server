package conf

import (
	"os"
	"os/user"
)

var BaseConf struct {
	Hostname string
	Username string

	PrjName string
	IsDev   bool

	TmpRoot    string
	ApiPidFile string

	UploadRoutineCnt int
}

func initBaseConf() {
	BaseConf.Hostname, _ = os.Hostname()
	curUser, _ := user.Current()
	BaseConf.Username = curUser.Username

	BaseConf.PrjName = scJson.PrjName
	BaseConf.IsDev = scJson.IsDev

	BaseConf.TmpRoot = PrjHome + "/tmp"
	BaseConf.ApiPidFile = BaseConf.TmpRoot + "/api.pid"

	BaseConf.UploadRoutineCnt = scJson.UploadRoutineCnt
}
