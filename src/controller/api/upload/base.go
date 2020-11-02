package upload

import (
	"code-sync-server/conf"
	"code-sync-server/controller/api"
	uploadSvc "code-sync-server/svc/upload"
	"github.com/goinbox/crypto"
	gcontroller "github.com/goinbox/gohttp/controller"
	"os"
	"strconv"

	"errors"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	MultiPartReadBufSize = 4096

	MultiPartFormNameMd5     = "md5"
	MultiPartFormNameFile    = "formfile"
	MultiPartFormNamePerm    = "perm"
	MultiPartFormNameVersion = "version"
)

type multiPart struct {
	fileName string
	contents []byte
}

type UploadContext struct {
	*api.ApiContext

	uploadSvc *uploadSvc.UploadSvc

	multiForm map[string]*multiPart
}

func (uc *UploadContext) BeforeAction() {
	uc.ApiContext.BeforeAction()

	uc.uploadSvc = uploadSvc.NewUploadSvc(uc.TraceId)
}

type UploadController struct {
	api.BaseController
}

func (uc *UploadController) NewActionContext(req *http.Request, respWriter http.ResponseWriter) gcontroller.ActionContext {
	context := new(UploadContext)
	context.ApiContext = uc.BaseController.NewActionContext(req, respWriter).(*api.ApiContext)
	context.multiForm = make(map[string]*multiPart)

	return context
}

func (uc *UploadController) parseMultipart(context *UploadContext) error {
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

		err = uc.parsePart(context, part)
		if err != nil {
			return err
		}
	}
}

func (uc *UploadController) parsePart(context *UploadContext, part *multipart.Part) error {
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

func (uc *UploadController) parseUploadFile(prj string, context *UploadContext) (*uploadSvc.UploadFile, map[string][]byte, error) {
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

	versionPart, ok := context.multiForm[MultiPartFormNameVersion]
	if !ok {
		return nil, nil, errors.New("not have file version")
	}

	originPartData[MultiPartFormNameVersion] = versionPart.contents
	versionStr := string(versionPart.contents)
	versionInt, err := strconv.Atoi(versionStr)
	if err != nil {
		return nil, nil, err
	}

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
	msg += "version:" + versionStr
	context.InfoLog([]byte("parseUploadFile"), []byte(msg))

	return &uploadSvc.UploadFile{
		Cpc:      cpc,
		Rpath:    contentsPart.fileName,
		Contents: contentsPart.contents,
		Perm:     perm,
		Version:  versionInt,
	}, originPartData, nil
}
