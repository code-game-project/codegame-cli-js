package main

import (
	"fmt"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/cggenevents"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/modules"
	"github.com/code-game-project/go-utils/server"
)

func Update(projectName string) error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	data, err := modules.ReadCommandConfig[modules.UpdateData]()
	if err != nil {
		return err
	}
	switch config.Type {
	case "client":
		return updateClient(projectName, data.LibraryVersion, config)
	case "server":
		return updateServer(data.LibraryVersion)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func updateClient(projectName, libraryVersion string, config *cgfile.CodeGameFileData) error {
	api, err := server.NewAPI(config.URL)
	if err != nil {
		return err
	}
	info, err := api.FetchGameInfo()
	if err != nil {
		return err
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

	typescript := config.Lang == "ts"

	runtime, _ := config.LangConfig["runtime"].(string)
	if runtime != "node" && runtime != "bundler" && (typescript || runtime != "browser") {
		return fmt.Errorf("Invalid runtime: '%s'", runtime)
	}

	err = updateClientTemplate(projectName, info, eventNames, commandNames, runtime, typescript)
	if err != nil {
		return err
	}

	cli.BeginLoading("Updating dependencies...")
	_, err = exec.Execute(true, "npm", "install", "@code-game-project/client"+"@"+libraryVersion)
	if err != nil {
		return err
	}
	switch runtime {
	case "browser":
		_, err = exec.Execute(true, "npm", "install", "--save-dev", "serve@latest")
	case "bundler":
		_, err = exec.Execute(true, "npm", "install", "--save-dev", "parcel@latest")
	}
	if err != nil {
		return err
	}
	if typescript {
		_, err = exec.Execute(true, "npm", "install", "--save-dev", "typescript@latest", "@types/node@latest")
		if err != nil {
			return err
		}
	}
	_, err = exec.Execute(true, "npm", "update")
	if err != nil {
		return err
	}
	cli.FinishLoading()
	return nil
}

func updateClientTemplate(projectName string, info server.GameInfo, eventNames, commandNames []string, runtime string, typescript bool) error {
	return execClientTemplate(projectName, info, eventNames, commandNames, runtime, typescript, true)
}

func updateServer(libraryVersion string) error {
	panic("not implemented")
	cli.Warn("This update might include breaking changes. You will have to manually update your code to work with the new version.")
	ok, err := cli.YesNo("Continue?", false)
	if err != nil || !ok {
		return cli.ErrCanceled
	}

	cli.BeginLoading("Updating dependencies...")
	cli.FinishLoading()
	return nil
}
