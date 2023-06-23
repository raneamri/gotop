package io

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"
)

func WriteConfig(instance types.Instance) error {
	/*
		Parse .conf to find connections section
	*/
	fpath := "gotop.conf"
	parser, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	connStart := strings.Index(string(parser), "[/connections]")
	connEnd := strings.Index(string(parser), "[/connections]")
	beforeSection := string(parser[:connStart])
	afterSection := string(parser[connEnd:])

	var (
		fdbms   string = "dbms=" + utility.Strdbms(instance.DBMS) + " "
		fdsn    string = "dsn=" + string(instance.DSN) + " "
		fdbname string = "conn-name=" + instance.ConnName

		parsed []byte = []byte(beforeSection + fdbms + fdsn + fdbname + "\n" + afterSection)
	)

	err = ioutil.WriteFile(fpath, parsed, 0644)
	if err != nil {
		return err
	}

	return nil
}

/*
Heals a file by iterating through its content and finding irregularities
Note: yet to be completed
*/
func HealConfig() {
	//fpath := "./gotop.conf"
	//parser, err := ioutil.ReadFile(fpath)

}

/*
Deletes the config file and recreates it, then rewrites headers
*/
func ResetConfig() {
	var (
		fpath   string
		headers [2]string
	)
	fpath = "gotop.conf"
	/*
		Recreate file
	*/
	os.Remove(fpath)
	os.Create(fpath)
	/*
		Re-write headers
	*/
	headers = [2]string{"plugins", "connections"}
	for _, elem := range headers {
		var parsed []byte = []byte("[" + elem + "]\n\n[/" + elem + "]\n")
		ioutil.WriteFile(fpath, parsed, 0644)
	}
}

/*
Reads preset instances and puts them in slice by default
*/
func ReadConfig(instances []types.Instance) ([]types.Instance, error) {
	/*
		Same parsing as writeConfig
	*/
	var (
		fpath     string
		connStart int
		connEnd   int
		config    string
	)

	fpath = "gotop.conf"
	contents, err := ioutil.ReadFile(fpath)
	if err != nil {
		return instances, err
	}

	/*
		Only keep 'connections' section
	*/
	connStart = strings.Index(string(contents), "[connections]")
	connEnd = strings.Index(string(contents), "[/connections]")
	config = string(contents[connStart+13 : connEnd-1])

	/*
		Parse each line
	*/
	var lines []string = strings.Split(config, "\n")
	for _, line := range lines {
		var (
			inst  types.Instance
			pairs []string
			parts []string
			key   string
			value string
		)

		line = strings.TrimSpace(line)
		if line == "" || line == "\n" {
			continue
		}

		pairs = strings.Split(line, " ")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			parts = strings.Split(pair, "=")
			if len(parts) != 2 {
				continue
			}

			key = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])

			switch key {
			case "dbms":
				inst.DBMS = utility.Dbmsstr(value)
			case "dsn":
				inst.DSN = []byte(value)
			case "conn-name":
				inst.ConnName = value
			}
		}

		instances = utility.PushInstance(instances, inst)
	}

	return instances, err
}

/*
Removes duplicates in config
*/
func CleanConfig() {
	var (
		fpath       string
		config      string
		output      string
		lines       []string
		uniquelines map[string]struct{}
		ordered     []string
	)
	fpath = "gotop.conf"
	parser, err := ioutil.ReadFile(fpath)
	if err != nil {
		panic(err)
	}
	config = string(parser)
	lines = strings.Split(config, "\n")
	uniquelines = make(map[string]struct{})

	for _, line := range lines {
		/*
			Skip empty lines
		*/
		if line == "" || line == "\n" {
			uniquelines[line] = struct{}{}
			ordered = append(ordered, line)
			continue
		}

		/*
			Check if line exists
		*/
		if _, exists := uniquelines[line]; exists {
			continue
		}

		/*
			Map line as unique
		*/
		uniquelines[line] = struct{}{}
		ordered = append(ordered, line)
	}

	/*
		Format the unique lines
	*/
	output = strings.Join(ordered, "\n")

	/*
		Re-write config
	*/
	err = ioutil.WriteFile(fpath, []byte(output), 0644)
	if err != nil {
		panic(err)
	}
}

/*
Syncs []Instance slice to config
Note: fix duplicates bug
*/
func SyncConfig(instances []types.Instance) []types.Instance {
	/*
		Write all instances in object to file
	*/
	for _, inst := range instances {
		err := WriteConfig(inst)
		if err != nil {
			panic(err)
		}
	}
	/*
		Remove duplicates accross object & file
	*/
	CleanConfig()
	/*
		Clean instances and read cleaned config
	*/
	instances = instances[:0]
	instances, _ = ReadConfig(instances)
	return instances
}
