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

const (
	ControlStop = 1

	OpUpload = 1
	OpDelete = 2
)

type SyncFile struct {
	Op       int
	Cpc      *conf.CodePrjConf
	Rpath    string
	Contents []byte
	Perm     os.FileMode
}

type SyncSvc struct {
	*svc.BaseSvc
}

var syncFileChanList []chan *SyncFile
var syncControlChanList []chan int

func StartSyncRoutine() {
	cnt := conf.BaseConf.SyncRoutineCnt

	syncFileChanList = make([]chan *SyncFile, cnt)
	syncControlChanList = make([]chan int, cnt)

	for i := 0; i < cnt; i++ {
		sfCh := make(chan *SyncFile)
		syncFileChanList[i] = sfCh

		scCh := make(chan int)
		syncControlChanList[i] = scCh

		ss := NewSyncSvc([]byte(strconv.Itoa(i)))
		go ss.SyncRoutine(sfCh, scCh)
	}

}

func StopSyncRoutine() {
	for i := 0; i < conf.BaseConf.SyncRoutineCnt; i++ {
		syncControlChanList[i] <- ControlStop
	}
}

func NewSyncSvc(traceId []byte) *SyncSvc {
	return &SyncSvc{
		BaseSvc: &svc.BaseSvc{
			TraceId: traceId,
		},
	}
}

func (ss *SyncSvc) SyncFile(sf *SyncFile) {
	i := ss.findRoutineIndex(sf)

	syncFileChanList[i] <- sf
}

func (ss *SyncSvc) findRoutineIndex(sf *SyncFile) int {
	str := crypto.Md5String([]byte(sf.Rpath))
	hv, _ := strconv.ParseInt(str[len(str)-1:], 16, 10)
	i := int(hv) % conf.BaseConf.SyncRoutineCnt

	return i
}

func (ss *SyncSvc) SyncRoutine(sfCh chan *SyncFile, scCh chan int) {
	for {
		select {
		case sf := <-sfCh:
			switch sf.Op {
			case OpUpload:
				ss.saveFile(sf)
			case OpDelete:
				ss.deleteFile(sf)
			default:
				ss.ErrorLog([]byte("SyncRoutine"), []byte("unknown op "+strconv.Itoa(sf.Op)))
			}
		case <-scCh:
			return
		}
	}
}

func (ss *SyncSvc) saveFile(sf *SyncFile) {
	path := sf.Cpc.PrjHome + "/" + sf.Rpath
	dir := filepath.Dir(path)
	if !gomisc.DirExist(dir) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			ss.ErrorLog([]byte("saveFile"), []byte(err.Error()))
			return
		}
	}

	exist := gomisc.FileExist(path)
	err := ioutil.WriteFile(path, sf.Contents, sf.Perm)
	if err != nil {
		ss.ErrorLog([]byte("saveFile"), []byte(err.Error()))
		return
	}

	if exist {
		err := os.Chmod(path, sf.Perm)
		if err != nil {
			ss.ErrorLog([]byte("chmod"), []byte(err.Error()))
		}
	}
}

func (ss *SyncSvc) deleteFile(sf *SyncFile) {
	path := sf.Cpc.PrjHome + "/" + sf.Rpath
	ss.InfoLog([]byte("deleteFile"), []byte(path))

	err := os.RemoveAll(path)
	if err != nil {
		ss.ErrorLog([]byte("deleteFile"), []byte(err.Error()))
	}
}
