// +build !pro

package crane

import (
	"fmt"
)

const Version = "3.3.2"
const Pro = false

func printVersion() {
	fmt.Printf("v%s\n", Version)
}
