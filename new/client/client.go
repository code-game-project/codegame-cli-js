package client

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli-js/new"
	"github.com/code-game-project/codegame-cli/util"
)

//go:embed templates/index-node.js.tmpl
var indexJSNode string

//go:embed templates/tsconfig.json
var tsConfig []byte

func CreateNewClient(projectName, serverURL, libraryVersion string, typescript bool) error {
	cli.BeginLoading("Installing javascript-client...")

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

	cli.FinishLoading()

	err = createTemplate(projectName, serverURL, typescript)
	if err != nil {
		return err
	}

	if typescript {
		cli.BeginLoading("Installing dependencies...")
		_, err = util.Execute(true, "npm", "install", "--save-dev", "@types/node")
		if err != nil {
			return err
		}
		cli.FinishLoading()
	}

	err = new.ConfigurePackageJSON()
	if err != nil {
		return err
	}

	return nil
}

func createTemplate(projectName, serverURL string, typescript bool) error {
	path := filepath.Join("src", "index.js")
	if typescript {
		path = filepath.Join("src", "index.ts")
	}

	err := executeTemplate(indexJSNode, path, projectName, serverURL)
	if err != nil {
		return err
	}

	if typescript {
		err = os.WriteFile("tsconfig.json", tsConfig, 0644)
		if err != nil {
			return err
		}
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
