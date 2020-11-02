package conf

import (
	"github.com/goinbox/color"

	"fmt"
	"reflect"
	"testing"
)

func init() {
	prjHome := "/Users/gibsonli/devspace/personal/code-sync-server"

	e := Init(prjHome)
	if e != nil {
		fmt.Println("Init error: ", e.Error())
	}
}

func TestConf(t *testing.T) {
	t.Log("PrjHome", PrjHome)
	printComplexObjectForTest(&BaseConf)
	printComplexObjectForTest(&LogConf)
	printComplexObjectForTest(&ApiHttpConf)

	for name, cpc := range CodePrjConfMap {
		t.Log(name)
		printComplexObjectForTest(cpc)
	}
	printComplexObjectForTest(&CodePrjConf{})
}

func printComplexObjectForTest(v interface{}) {
	vo := reflect.ValueOf(v)
	elems := vo.Elem()
	ts := elems.Type()

	c := color.Yellow([]byte("Print detail: "))
	fmt.Println(string(c), vo.Type())
	for i := 0; i < elems.NumField(); i++ {
		field := elems.Field(i)
		fmt.Println(ts.Field(i).Name, field.Type(), field.Interface())
	}
}
