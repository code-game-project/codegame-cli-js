package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/cggenevents"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/modules"
	"github.com/code-game-project/go-utils/server"
)

//go:embed templates/new/client/package.json.tmpl
var clientPackageJSONTemplate string

//go:embed templates/new/client/js/app.js.tmpl
var clientJSIndexTemplate string

//go:embed templates/new/client/js/game.js.tmpl
var clientJSGameTemplate string

//go:embed templates/new/client/ts/tsconfig.json.tmpl
var clientTSConfigTemplate string

//go:embed templates/new/client/ts/index.ts.tmpl
var clientTSIndexTemplate string

//go:embed templates/new/client/ts/game.ts.tmpl
var clientTSGameTemplate string

//go:embed templates/new/client/js/index.html.tmpl
var clientIndexHTMLTemplate string

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

	cgConf, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}
	cgConf.LangConfig["runtime"] = runtime
	err = cgConf.Write("")
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
		if typescript {
			_, err = exec.Execute(true, "npm", "install", "--save-dev", "typescript", "@types/node")
			if err != nil {
				return err
			}
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
		cli.Warn("This action will ERASE and regenerate ALL files in '%s/'.\nYou will have to manually update your code to work with the new version.", info.Name)
		ok, err := cli.YesNo("Continue?", false)
		if err != nil || !ok {
			return cli.ErrCanceled
		}
		os.RemoveAll(info.Name)
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
			err := ExecTemplate(clientJSIndexTemplate, "app.js", data)
			if err != nil {
				return err
			}
		}
		err := ExecTemplate(clientJSGameTemplate, filepath.Join(info.Name, "game.js"), data)
		if err != nil {
			return err
		}
		if !node {
			err := ExecTemplate(clientIndexHTMLTemplate, "index.html", data)
			if err != nil {
				return err
			}
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
