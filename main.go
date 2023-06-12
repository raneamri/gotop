package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/raneamri/gotop/errors"
	"github.com/raneamri/gotop/io"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/ui"
	"github.com/raneamri/gotop/utility"
)

/*
View concurrencies
View plugins
View make file
*/

func main() {

	utility.ClearTerminal()

	/*
		Slice to store all instances
	*/
	var (
		instances []types.Instance
		err       error
	)
	/*
		Attempt to fetch config from .conf
	*/
	instances, err = io.ReadConfig()
	if err != nil {
		errors.CatchConfigReadError(err, instances)
	}
	/*
		Fetch configs from prompt/args
	*/
	instances = io.ReadArgs(instances)

	ui.InterfaceLoop(instances)
}

func Version() string {
	return "v0.0.0"
}
