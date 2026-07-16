// Command kvstore runs the comparative capstone.
package main

import (
	"context"
	"os"

	"github.com/mbrndiar/learning-go/capstones/comparative/solution/kvstore/app"
)

func main() {
	os.Exit(app.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}
