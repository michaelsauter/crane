package crane

import (
	"fmt"
)

const Version = "3.5.0"
const Pro = false

func printVersion() {
	fmt.Printf("v%s\n", Version)
}
