package crane

import (
	"fmt"
)

const Version = "3.6.1"

func printVersion() {
	fmt.Printf("v%s\n", Version)
}
