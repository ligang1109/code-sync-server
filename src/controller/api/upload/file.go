package upload

import (
	"github.com/goinbox/goerror"
	"github.com/goinbox/gohttp/query"

	"code-sync-server/errno"
	"code-sync-server/misc"
)

type fileActionParams struct {
	prj  string
	user string
	host string

	misc.ApiSignParams
}

var fileSignQueryNames = append([]string{"prj", "user", "host", MultiPartFormNameFile, MultiPartFormNameMd5, MultiPartFormNamePerm, MultiPartFormNameVersion}, misc.ApiSignQueryNames...)

func (uc *UploadController) FileAction(context *UploadContext) {
	ap, _, e := uc.parseFileActionParams(context)
	if e != nil {
		context.ApiData.Err = e
		return
	}

	err := uc.parseMultipart(context)
	if err != nil {
		context.ApiData.Err = goerror.New(errno.EParseMultipartError, err.Error())
		return
	}

	uf, opd, err := uc.parseUploadFile(ap.prj, context)
	if err != nil {
		context.ApiData.Err = goerror.New(errno.EParseUploadFileError, err.Error())
		return
	}

	context.QueryValues.Set(MultiPartFormNameFile, uf.Rpath)
	context.QueryValues.Set(MultiPartFormNameMd5, string(opd[MultiPartFormNameMd5]))
	context.QueryValues.Set(MultiPartFormNamePerm, string(opd[MultiPartFormNamePerm]))
	context.QueryValues.Set(MultiPartFormNameVersion, string(opd[MultiPartFormNameVersion]))
	e = misc.VerifyApiSign(&ap.ApiSignParams, context.QueryValues, fileSignQueryNames, uf.Cpc.Token)
	if e != nil {
		context.ApiData.Err = e
		return
	}

	context.uploadSvc.UploadFile(uf)
}

func (uc *UploadController) parseFileActionParams(context *UploadContext) (*fileActionParams, map[string]bool, *goerror.Error) {
	ap := new(fileActionParams)

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
