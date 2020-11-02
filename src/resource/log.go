package resource

import (
	"code-sync-server/conf"
	"code-sync-server/errno"

	"github.com/goinbox/goerror"
	"github.com/goinbox/golog"
)

var accessLogWriter golog.IWriter
var AccessLogger golog.ILogger

var traceLogWriter golog.IWriter
var TraceLogger golog.ILogger

func InitLog(systemName string) *goerror.Error {
	if conf.BaseConf.IsDev {
		accessLogWriter = golog.NewConsoleWriter()
	} else {
		fw, err := golog.NewFileWriter(conf.LogConf.RootPath+"/"+systemName+"_access.log", conf.LogConf.Bufsize)
		if err != nil {
			return goerror.New(errno.ESysInitLogFail, err.Error())
		}
		accessLogWriter = golog.NewAsyncWriter(fw, conf.LogConf.AsyncQueueSize)
	}
	AccessLogger = NewLogger(accessLogWriter)

	fw, err := golog.NewFileWriter(conf.LogConf.RootPath+"/"+systemName+"_trace.log", conf.LogConf.Bufsize)
	if err != nil {
		return goerror.New(errno.ESysInitLogFail, err.Error())
	}
	traceLogWriter = golog.NewAsyncWriter(fw, conf.LogConf.AsyncQueueSize)
	TraceLogger = NewLogger(traceLogWriter)

	return nil
}

func NewLogger(writer golog.IWriter) golog.ILogger {
	return golog.NewSimpleLogger(writer, golog.NewSimpleFormater()).SetLogLevel(conf.LogConf.Level)
}

func FreeLog() {
	accessLogWriter.Free()
	traceLogWriter.Free()
}
