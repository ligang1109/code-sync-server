package file

import (
	"github.com/goinbox/goerror"
	"github.com/goinbox/gohttp/query"

	"code-sync-server/errno"
	"code-sync-server/misc"
)

type uploadActionParams struct {
	prj  string
	user string
	host string

	misc.ApiSignParams
}

var uploadSignQueryNames = append([]string{"prj", "user", "host", MultiPartFormNameFile, MultiPartFormNameMd5, MultiPartFormNamePerm}, misc.ApiSignQueryNames...)

func (fc *FileController) UploadAction(context *FileContext) {
	ap, _, e := fc.parseUploadActionParams(context)
	if e != nil {
		context.ApiData.Err = e
		return
	}

	err := fc.parseMultipart(context)
	if err != nil {
		context.ApiData.Err = goerror.New(errno.EParseMultipartError, err.Error())
		return
	}

	uf, opd, err := fc.parseUploadFile(ap.prj, context)
	if err != nil {
		context.ApiData.Err = goerror.New(errno.EParseUploadFileError, err.Error())
		return
	}

	context.QueryValues.Set(MultiPartFormNameFile, uf.Rpath)
	context.QueryValues.Set(MultiPartFormNameMd5, string(opd[MultiPartFormNameMd5]))
	context.QueryValues.Set(MultiPartFormNamePerm, string(opd[MultiPartFormNamePerm]))
	e = misc.VerifyApiSign(&ap.ApiSignParams, context.QueryValues, uploadSignQueryNames, uf.Cpc.Token)
	if e != nil {
		context.ApiData.Err = e
		return
	}

	context.syncSvc.SyncFile(uf)
}

func (fc *FileController) parseUploadActionParams(context *FileContext) (*uploadActionParams, map[string]bool, *goerror.Error) {
	ap := new(uploadActionParams)

	qs := query.NewQuerySet()
	qs.StringVar(&ap.prj, "prj", true, errno.ECommonInvalidArg, "invalid prj", query.CheckStringNotEmpty)
	qs.StringVar(&ap.user, "user", true, errno.ECommonInvalidArg, "invalid user", query.CheckStringNotEmpty)
	qs.StringVar(&ap.host, "host", true, errno.ECommonInvalidArg, "invalid host", query.CheckStringNotEmpty)

	misc.SetApiSignParams(qs, &ap.ApiSignParams)

	e := qs.Parse(context.QueryValues)
	if e != nil {
		return ap, nil, e
	}

	return ap, qs.ExistsInfo(), nil
}
