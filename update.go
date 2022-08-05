package main

import (
	"fmt"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/modules"
	"github.com/code-game-project/go-utils/server"
)

func Update() error {
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
		return updateClient(data.LibraryVersion, config)
	case "server":
		return updateServer(data.LibraryVersion)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func updateClient(libraryVersion string, config *cgfile.CodeGameFileData) error {
	return nil
}

func updateClientTemplate() error {
	return execClientTemplate("", server.GameInfo{}, nil, nil, false, false, true)
}

func updateServer(libraryVersion string) error {
	cli.Warn("This update might include breaking changes. You will have to manually update your code to work with the new version.")
	ok, err := cli.YesNo("Continue?", false)
	if err != nil || !ok {
		return cli.ErrCanceled
	}

	cli.BeginLoading("Updating dependencies...")
	cli.FinishLoading()
	return nil
}
