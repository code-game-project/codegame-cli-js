package server

import (
	_ "embed"
	"os"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli-js/new"
	"github.com/code-game-project/codegame-cli/util"
)

//go:embed templates/tsconfig.json
var tsConfig []byte

func CreateNewServer(projectName, libraryVersion string) error {
	cli.Warn("This feature is not fully implemented yet.")

	cli.BeginLoading("Installing javascript-server...")

	var version string
	var err error
	if libraryVersion == "latest" {
		version = "latest"
	} else {
		version, err = new.NPMVersion("@code-game-project", "javascript-server", libraryVersion)
		if err != nil {
			return err
		}
	}

	_, err = util.Execute(true, "npm", "install", "@code-game-project/javascript-client"+"@", version)
	if err != nil {
		return err
	}

	cli.FinishLoading()

	err = createTemplate(projectName)
	if err != nil {
		return err
	}

	cli.BeginLoading("Installing dependencies...")
	_, err = util.Execute(true, "npm", "install", "--save-dev", "@types/node")
	if err != nil {
		return err
	}
	cli.FinishLoading()

	err = new.ConfigurePackageJSON()
	if err != nil {
		return err
	}

	return nil
}

func createTemplate(projectName string) error {
	err := os.WriteFile("tsconfig.json", tsConfig, 0644)
	if err != nil {
		return err
	}

	return nil
}

func executeTemplate(templateText, fileName, projectName string) error {
	type data struct {
		Name string
	}

	return new.ExecTemplate(templateText, fileName, data{
		Name: projectName,
	})
}
