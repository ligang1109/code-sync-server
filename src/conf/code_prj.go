package conf

import "strings"

type codePrjConfJson struct {
	PrjHome string `json:"prj_home"`
	Token   string `json:"token"`
}

type CodePrjConf struct {
	PrjName string
	PrjHome string
	Token   string
}

var CodePrjConfMap map[string]*CodePrjConf

func initCodePrjConf() {
	CodePrjConfMap = make(map[string]*CodePrjConf)
	for name, item := range scJson.CodePrjMap {
		CodePrjConfMap[name] = &CodePrjConf{
			PrjName: name,
			PrjHome: strings.TrimRight(item.PrjHome, "/"),
			Token:   item.Token,
		}
	}
}
