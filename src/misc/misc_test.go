package misc

import (
	"flag"
	"testing"
	"time"
)

func TestStructSimpleFieldAssign(t *testing.T) {
	type ta struct {
		Id   int64
		Name string
	}
	type tb struct {
		Id   time.Duration
		Name string
	}

	a := &ta{
		Id:   10,
		Name: "a",
	}
	b := new(tb)

	StructSimpleFieldAssign(a, b)
	t.Log(a, b)
}

//go test -v -test.run=TestGenPrjToken -args "xxxxxx"
func TestGenPrjToken(t *testing.T) {
	flag.Parse()

	args := flag.Args()
	prj := "demo"
	if len(args) > 0 {
		prj = args[0]
	}

	t.Log(GenPrjToken(prj))
}
