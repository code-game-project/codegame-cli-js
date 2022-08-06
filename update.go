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

	err = updateClientTemplate(projectName, info, eventNames, commandNames, config.LangConfig["runtime"] == "node", config.Lang == "ts")
	if err != nil {
		return err
	}

	cli.BeginLoading("Updating dependencies...")
	_, err = exec.Execute(true, "npm", "install", "@code-game-project/client"+"@"+libraryVersion)
	if err != nil {
		return err
	}
	_, err = exec.Execute(true, "npm", "update")
	if err != nil {
		return err
	}
	cli.FinishLoading()
	return nil
}

func updateClientTemplate(projectName string, info server.GameInfo, eventNames, commandNames []string, node, typescript bool) error {
	return execClientTemplate(projectName, info, eventNames, commandNames, node, typescript, true)
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
