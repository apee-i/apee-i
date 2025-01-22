package cmd

import (
	"fmt"
)

// Update function updates the utility to the latest version
func Update() {

}

// Help function lists down all the commands and flags with their functionalities
func Help() {
	subCommands := map[string]string{
		"update": "\t - Updates the `apee-i` utility to the latest version",
	}

	fmt.Print("\n\tSUBCOMMANDS\n\n")

	for key, value := range subCommands {
		fmt.Println(key, value)
	}

	flags:= map[string]string{
		"-help/--help": "\t\t - lists all the options within the application",
		"-file/--file": "\t\t - enter file path(json/yaml). Default file name is api.json",
		"-env/--env": "\t\t - enter env (development/staging/production). Default is development",
		"-pipeline/--pipeline": "\t - enter pipeline type (current/custom/all). Default is current",
		"-name/--name": "\t\t - enter custom pipeline name (names defined in your custom pipelines section)\n\t\t\t   Only works if --pipeline flag is set to custom like so --pipeline=custom or -pipeline=custom",
	}

	fmt.Print("\n\tFLAGS\n\n")

	for key, value := range flags {
		fmt.Println(key, value)
	}
}

// Version fetches and displays the current running version of the Application
func Version() {
	fmt.Println("1.0.0")
}
