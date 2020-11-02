package svc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/goinbox/crypto"
	"github.com/goinbox/gomisc"

	"code-sync-server/conf"
	"code-sync-server/svc"
)

const ControlStop = 1

type UploadFile struct {
	Cpc      *conf.CodePrjConf
	Rpath    string
	Contents []byte
	Perm     os.FileMode
	Version  int
}

type UploadSvc struct {
	*svc.BaseSvc
}

var uploadFileChanList []chan *UploadFile
var uploadControlChanList []chan int

func StartUploadRoutine() {
	cnt := conf.BaseConf.UploadRoutineCnt

	uploadFileChanList = make([]chan *UploadFile, cnt)
	uploadControlChanList = make([]chan int, cnt)

	for i := 0; i < cnt; i++ {
		ufCh := make(chan *UploadFile)
		uploadFileChanList[i] = ufCh

		ucCh := make(chan int)
		uploadControlChanList[i] = ucCh

		us := NewUploadSvc([]byte(strconv.Itoa(i)))
		go us.UploadRoutine(ufCh, ucCh)
	}

}

func StopUploadRoutine() {
	for i := 0; i < conf.BaseConf.UploadRoutineCnt; i++ {
		uploadControlChanList[i] <- ControlStop
	}
}

func NewUploadSvc(traceId []byte) *UploadSvc {
	return &UploadSvc{
		BaseSvc: &svc.BaseSvc{
			TraceId: traceId,
		},
	}
}

func (us *UploadSvc) UploadFile(uf *UploadFile) {
	i := us.findRoutineIndex(uf)

	uploadFileChanList[i] <- uf
}

func (us *UploadSvc) findRoutineIndex(uf *UploadFile) int {
	str := crypto.Md5String([]byte(uf.Rpath))
	hv, _ := strconv.ParseInt(str[len(str)-1:], 16, 10)
	i := int(hv) % conf.BaseConf.UploadRoutineCnt

	return i
}

func (us *UploadSvc) UploadRoutine(ufCh chan *UploadFile, ucCh chan int) {
	for {
		select {
		case uf := <-ufCh:
			us.saveFile(uf)
		case <-ucCh:
			return
		}
	}
}

func (us *UploadSvc) saveFile(uf *UploadFile) {
	path := uf.Cpc.PrjHome + "/" + uf.Rpath
	dir := filepath.Dir(path)
	if !gomisc.DirExist(dir) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			us.ErrorLog([]byte("saveFile"), []byte(err.Error()))
			return
		}
	}

	exist := gomisc.FileExist(path)
	err := ioutil.WriteFile(path, uf.Contents, uf.Perm)
	if err != nil {
		us.ErrorLog([]byte("saveFile"), []byte(err.Error()))
		return
	}

	if exist {
		err := os.Chmod(path, uf.Perm)
		if err != nil {
			us.ErrorLog([]byte("chmod"), []byte(err.Error()))
		}
	}
}
