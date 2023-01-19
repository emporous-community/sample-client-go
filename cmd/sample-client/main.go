package main

import (
	"github.com/spf13/cobra"

	"github.com/emporous-community/sample-client-go/cmd/sample-client/commands"
)

func main() {
	app := commands.NewRootCmd()
	cobra.CheckErr(app.Execute())
}
