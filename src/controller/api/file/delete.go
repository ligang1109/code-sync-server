package file

import (
	"code-sync-server/conf"
	syncSvc "code-sync-server/svc/sync"
	"github.com/goinbox/goerror"
	"github.com/goinbox/gohttp/query"

	"code-sync-server/errno"
	"code-sync-server/misc"
)

type deleteActionParams struct {
	prj   string
	user  string
	host  string
	rpath string

	misc.ApiSignParams
}

var deleteSignQueryNames = append([]string{"prj", "user", "host", "rpath"}, misc.ApiSignQueryNames...)

func (fc *FileController) DeleteAction(context *FileContext) {
	ap, _, e := fc.parseDeleteActionParams(context)
	if e != nil {
		context.ApiData.Err = e
		return
	}

	cpc, ok := conf.CodePrjConfMap[ap.prj]
	if !ok {
		context.ApiData.Err = goerror.New(errno.ECommonInvalidArg, "prj not exist")
		return
	}

	e = misc.VerifyApiSign(&ap.ApiSignParams, context.QueryValues, deleteSignQueryNames, cpc.Token)
	if e != nil {
		context.ApiData.Err = e
		return
	}

	context.syncSvc.SyncFile(&syncSvc.SyncFile{
		Op:    syncSvc.OpDelete,
		Cpc:   cpc,
		Rpath: ap.rpath,
	})
}

func (fc *FileController) parseDeleteActionParams(context *FileContext) (*deleteActionParams, map[string]bool, *goerror.Error) {
	ap := new(deleteActionParams)

	qs := query.NewQuerySet()
	qs.StringVar(&ap.prj, "prj", true, errno.ECommonInvalidArg, "invalid prj", query.CheckStringNotEmpty)
	qs.StringVar(&ap.user, "user", true, errno.ECommonInvalidArg, "invalid user", query.CheckStringNotEmpty)
	qs.StringVar(&ap.host, "host", true, errno.ECommonInvalidArg, "invalid host", query.CheckStringNotEmpty)
	qs.StringVar(&ap.rpath, "rpath", true, errno.ECommonInvalidArg, "invalid rpath", query.CheckStringNotEmpty)

	misc.SetApiSignParams(qs, &ap.ApiSignParams)

	e := qs.Parse(context.QueryValues)
	if e != nil {
		return ap, nil, e
	}

	return ap, qs.ExistsInfo(), nil
}
