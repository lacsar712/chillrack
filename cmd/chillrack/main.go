package main

import (
	"os"

	"github.com/lacsar712/chillrack/internal/app"
)

func main() {
	os.Exit(app.RunCLI())
}
