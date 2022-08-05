package main

import (
	_ "embed"
	"path/filepath"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cggenevents"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/modules"
	"github.com/code-game-project/go-utils/server"
)

//go:embed templates/new/client/package.json.tmpl
var clientPackageJSONTemplate string

//go:embed templates/new/client/index.js.tmpl
var clientJSIndexTemplate string

//go:embed templates/new/client/game.js.tmpl
var clientJSGameTemplate string

//go:embed templates/new/client/tsconfig.json.tmpl
var clientTSConfigTemplate string

//go:embed templates/new/client/index.ts.tmpl
var clientTSIndexTemplate string

//go:embed templates/new/client/game.ts.tmpl
var clientTSGameTemplate string

func CreateNewClient(projectName string) error {
	data, err := modules.ReadCommandConfig[modules.NewClientData]()
	if err != nil {
		return err
	}
	typescript := data.Lang == "ts"

	api, err := server.NewAPI(data.URL)
	if err != nil {
		return err
	}

	info, err := api.FetchGameInfo()
	if err != nil {
		return err
	}

	runtime, err := cli.SelectString("Runtime:", []string{"Browser", "Node.js"}, []string{"browser", "node"})
	if err != nil {
		return err
	}
	node := runtime == "node"

	if !node {
		panic("not implemented")
	}

	cge, err := api.GetCGEFile()
	if err != nil {
		return err
	}
	cgeVersion, err := cggenevents.ParseCGEVersion(cge)
	if err != nil {
		return err
	}

	eventNames, commandNames, err := cggenevents.GetEventNames(api.BaseURL(), cgeVersion)
	if err != nil {
		return err
	}

	err = createClientTemplate(projectName, info, eventNames, commandNames, node, typescript)
	if err != nil {
		return err
	}

	if node {
		cli.BeginLoading("Installing javascript-client...")
		_, err = exec.Execute(true, "npm", "install", "@code-game-project/client"+"@"+data.LibraryVersion)
		if err != nil {
			return err
		}
		cli.FinishLoading()

		cli.BeginLoading("Installing dependencies...")
		_, err = exec.Execute(true, "npm", "install", "commander")
		if err != nil {
			return err
		}
		_, err = exec.Execute(true, "npm", "install", "--save-dev", "typescript", "@types/node")
		if err != nil {
			return err
		}
		cli.FinishLoading()
	}

	return nil
}

func createClientTemplate(projectName string, info server.GameInfo, eventNames, commandNames []string, node, typescript bool) error {
	return execClientTemplate(projectName, info, eventNames, commandNames, node, typescript, false)
}

func execClientTemplate(projectName string, info server.GameInfo, eventNames, commandNames []string, node, typescript, update bool) error {
	if update {
		panic("not implemented")
	} else {
		cli.Warn("DO NOT EDIT the `%s/` directory inside of the project. ALL CHANGES WILL BE LOST when running `codegame update`.", info.Name)
	}

	type event struct {
		Name       string
		PascalName string
	}

	events := make([]event, len(eventNames))
	for i, e := range eventNames {
		pascal := strings.ReplaceAll(e, "_", " ")
		pascal = strings.Title(pascal)
		pascal = strings.ReplaceAll(pascal, " ", "")
		events[i] = event{
			Name:       e,
			PascalName: pascal,
		}
	}

	commands := make([]event, len(commandNames))
	for i, c := range commandNames {
		pascal := strings.ReplaceAll(c, "_", " ")
		pascal = strings.Title(pascal)
		pascal = strings.ReplaceAll(pascal, " ", "")
		commands[i] = event{
			Name:       c,
			PascalName: pascal,
		}
	}

	data := struct {
		ProjectName string
		GameName    string
		Version     string
		Node        bool
		TypeScript  bool
		Events      []event
		Commands    []event
	}{
		ProjectName: projectName,
		GameName:    info.Name,
		Version:     info.Version,
		Node:        node,
		TypeScript:  typescript,
		Events:      events,
		Commands:    commands,
	}

	if typescript {
		if !update {
			err := ExecTemplate(clientTSIndexTemplate, "src/index.ts", data)
			if err != nil {
				return err
			}
			err = ExecTemplate(clientTSConfigTemplate, "tsconfig.json", data)
			if err != nil {
				return err
			}
		}
		err := ExecTemplate(clientTSGameTemplate, filepath.Join("src", info.Name, "game.ts"), data)
		if err != nil {
			return err
		}
	} else {
		if !update {
			err := ExecTemplate(clientJSIndexTemplate, "src/index.js", data)
			if err != nil {
				return err
			}
		}
		err := ExecTemplate(clientJSGameTemplate, filepath.Join("src", info.Name, "game.js"), data)
		if err != nil {
			return err
		}
	}

	if !update && node {
		err := ExecTemplate(clientPackageJSONTemplate, "package.json", data)
		if err != nil {
			return err
		}
	}

	return nil
}
