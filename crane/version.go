package crane

import (
	"fmt"
)

const Version = "3.5.0"

func printVersion() {
	fmt.Printf("v%s\n", Version)
}
