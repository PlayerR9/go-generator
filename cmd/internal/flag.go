package internal

import "github.com/PlayerR9/go-generator"

var (
	// OutputFlag is the flag that specifies the output location of the generated code.
	OutputFlag *generator.OutputLocVal
)

func init() {
	OutputFlag = generator.NewOutputFlag("", false)
}
