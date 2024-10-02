package internal

import "github.com/PlayerR9/go-generator"

type MyData struct {
	PackageName string
}

func (d *MyData) SetPackageName(name string) {
	panic("LOL")
}

var (
	MyGenerator *generator.CodeGenerator[*MyData]
)

func init() {
	var err error
	MyGenerator, err = generator.NewCodeGeneratorFromTemplate[*MyData]("", templ)
	if err != nil {
		panic()
	}
}

var templ string = `
package {{.PackageName}}`
