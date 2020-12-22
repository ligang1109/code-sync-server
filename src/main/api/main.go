package main

import (
	"code-sync-server/conf"
	"code-sync-server/controller/api/file"
	"code-sync-server/errno"
	"code-sync-server/resource"
	syncSvc "code-sync-server/svc/sync"

	"github.com/goinbox/gohttp/gracehttp"
	"github.com/goinbox/gohttp/router"
	"github.com/goinbox/gohttp/system"
	"github.com/goinbox/pidfile"

	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
)

func main() {
	var prjHome string

	flag.StringVar(&prjHome, "prj-home", "", "prj-home absolute path")
	flag.Parse()

	prjHome = strings.TrimRight(prjHome, "/")
	if prjHome == "" {
		fmt.Println("missing flag prj-home: ")
		flag.PrintDefaults()
		os.Exit(errno.ESysInvalidPrjHome)
	}

	e := conf.Init(prjHome)
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(e.Errno())
	}

	e = resource.InitLog("api")
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(e.Errno())
	}
	defer func() {
		resource.FreeLog()
	}()

	syncSvc.StartSyncRoutine()
	defer func() {
		syncSvc.StopSyncRoutine()
	}()

	pf, err := pidfile.CreatePidFile(conf.BaseConf.ApiPidFile)
	if err != nil {
		fmt.Printf("create pid file %s failed, error: %s\n", conf.BaseConf.ApiPidFile, err.Error())
		os.Exit(errno.ESysSavePidFileFail)
	}

	r := router.NewSimpleRouter()
	r.MapRouteItems(
		new(file.FileController),
	)

	sys := system.NewSystem(r)

	err = gracehttp.ListenAndServe(conf.ApiHttpConf.GoHttpHost+":"+conf.ApiHttpConf.GoHttpPort, sys)
	if err != nil {
		fmt.Println("pid:" + strconv.Itoa(os.Getpid()) + ", err:" + err.Error())
	}

	if err := pidfile.ClearPidFile(pf); err != nil {
		fmt.Printf("clear pid file failed, error: %s\n", err.Error())
	}
}
