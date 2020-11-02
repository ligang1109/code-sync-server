package conf

import (
	"github.com/goinbox/gomisc"
)

var scJson serverConfJson

type serverConfJson struct {
	PrjName string `json:"prj_name"`
	IsDev   bool   `json:"is_dev"`

	Log     logConfJson  `json:"log"`
	ApiHttp httpConfJson `json:"api_http"`

	CodePrjMap map[string]*codePrjConfJson `json:"code_prj_map"`

	UploadRoutineCnt int `json:"upload_routine_cnt"`
}

func initServerConfJson() error {
	confRoot := PrjHome + "/conf"
	err := gomisc.ParseJsonFile(confRoot+"/server/server_conf.json", &scJson)
	if err != nil {
		return err
	}
	err = gomisc.ParseJsonFile(confRoot+"/server_conf_rewrite.json", &scJson)
	if err != nil {
		return err
	}

	return nil
}
