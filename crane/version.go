package crane

import (
	"fmt"
)

const Version = "3.6.0"

func printVersion() {
	fmt.Printf("v%s\n", Version)
}
