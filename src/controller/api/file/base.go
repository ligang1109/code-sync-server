package file

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/goinbox/crypto"
	gcontroller "github.com/goinbox/gohttp/controller"

	"code-sync-server/conf"
	"code-sync-server/controller/api"
	syncSvc "code-sync-server/svc/sync"
)

const (
	MultiPartReadBufSize = 4096

	MultiPartFormNameMd5  = "md5"
	MultiPartFormNameFile = "formfile"
	MultiPartFormNamePerm = "perm"
)

type multiPart struct {
	fileName string
	contents []byte
}

type FileContext struct {
	*api.ApiContext

	syncSvc *syncSvc.SyncSvc

	multiForm map[string]*multiPart
}

func (fc *FileContext) BeforeAction() {
	fc.ApiContext.BeforeAction()

	fc.syncSvc = syncSvc.NewSyncSvc(fc.TraceId)
}

type FileController struct {
	api.BaseController
}

func (fc *FileController) NewActionContext(req *http.Request, respWriter http.ResponseWriter) gcontroller.ActionContext {
	context := new(FileContext)
	context.ApiContext = fc.BaseController.NewActionContext(req, respWriter).(*api.ApiContext)
	context.multiForm = make(map[string]*multiPart)

	return context
}

func (fc *FileController) parseMultipart(context *FileContext) error {
	reader, err := context.Request().MultipartReader()
	if err != nil {
		return err
	}

	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		}

		err = fc.parsePart(context, part)
		if err != nil {
			return err
		}
	}
}

func (fc *FileController) parsePart(context *FileContext, part *multipart.Part) error {
	var contents []byte

	for {
		buf := make([]byte, MultiPartReadBufSize)
		n, err := part.Read(buf)

		if n > 0 {
			contents = append(contents, buf[0:n]...)
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
	}

	context.multiForm[part.FormName()] = &multiPart{
		fileName: part.FileName(),
		contents: contents,
	}

	return nil
}

func (fc *FileController) parseUploadFile(prj string, context *FileContext) (*syncSvc.SyncFile, map[string][]byte, error) {
	cpc, ok := conf.CodePrjConfMap[prj]
	if !ok {
		return nil, nil, errors.New("prj not exist")
	}

	permPart, ok := context.multiForm[MultiPartFormNamePerm]
	if !ok {
		return nil, nil, errors.New("not have file perm")
	}

	originPartData := make(map[string][]byte)

	originPartData[MultiPartFormNamePerm] = permPart.contents
	permInt, err := strconv.Atoi(string(permPart.contents))
	if err != nil {
		return nil, nil, err
	}
	perm := os.FileMode(permInt)

	md5Part, ok := context.multiForm[MultiPartFormNameMd5]
	if !ok {
		return nil, nil, errors.New("not have file md5")
	}

	originPartData[MultiPartFormNameMd5] = md5Part.contents
	contentsPart, ok := context.multiForm[MultiPartFormNameFile]
	if !ok {
		return nil, nil, errors.New("not have file contents")
	}

	originPartData[MultiPartFormNameFile] = contentsPart.contents
	if crypto.Md5String(contentsPart.contents) != string(md5Part.contents) {
		return nil, nil, errors.New("file md5 not equal")
	}

	msg := "rpath:" + contentsPart.fileName + "|"
	msg += "perm:" + perm.String() + "|"
	context.InfoLog([]byte("parseUploadFile"), []byte(msg))

	return &syncSvc.SyncFile{
		Op:       syncSvc.OpUpload,
		Cpc:      cpc,
		Rpath:    contentsPart.fileName,
		Contents: contentsPart.contents,
		Perm:     perm,
	}, originPartData, nil
}
