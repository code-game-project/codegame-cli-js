package client

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/code-game-project/codegame-cli-js/new"
	"github.com/code-game-project/codegame-cli/cli"
	"github.com/code-game-project/codegame-cli/util"
)

//go:embed templates/index-node.ts.tmpl
var indexTSNode string

//go:embed templates/tsconfig.json
var tsConfig []byte

func CreateNewClient(projectName, serverURL, libraryVersion string) error {
	cli.Begin("Installing correct javascript-client version...")

	var version string
	var err error
	if libraryVersion == "latest" {
		version = "latest"
	} else {
		version, err = new.NPMVersion("@code-game-project", "javascript-client", libraryVersion)
		if err != nil {
			return err
		}
	}

	_, err = util.Execute(true, "npm", "install", "@code-game-project/javascript-client"+"@"+version)
	if err != nil {
		return err
	}

	cli.Finish()

	cli.Begin("Creating project template...")
	err = createTemplate(projectName, serverURL)
	if err != nil {
		return err
	}
	cli.Finish()

	cli.Begin("Installing dependencies...")
	_, err = util.Execute(true, "npm", "install", "--save-dev", "@types/node")
	if err != nil {
		return err
	}
	cli.Finish()

	cli.Begin("Writing configuration files...")
	err = new.ConfigurePackageJSON()
	if err != nil {
		return err
	}
	cli.Finish()

	return nil
}

func createTemplate(projectName, serverURL string) error {
	err := executeTemplate(indexTSNode, filepath.Join("src", "index.ts"), projectName, serverURL)
	if err != nil {
		return err
	}

	err = os.WriteFile("tsconfig.json", tsConfig, 0644)
	if err != nil {
		return err
	}

	return nil
}

func executeTemplate(templateText, fileName, projectName, serverURL string) error {
	type data struct {
		Name string
		URL  string
	}

	return new.ExecTemplate(templateText, fileName, data{
		Name: projectName,
		URL:  serverURL,
	})
}
