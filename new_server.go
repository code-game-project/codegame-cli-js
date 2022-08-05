package main

import (
	"os"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/modules"

	_ "embed"
)

//go:embed templates/new/server/Dockerfile.tmpl
var serverDockerfileTemplate string

//go:embed templates/new/server/dockerignore.tmpl
var serverDockerignoreTemplate string

func CreateNewServer(projectName string) error {
	cli.Error("not implemented")
	os.Exit(1)

	data, err := modules.ReadCommandConfig[modules.NewServerData]()
	_ = data

	cli.BeginLoading("Installing js-server...")
	cli.FinishLoading()

	err = createServerTemplate(projectName)
	if err != nil {
		return err
	}

	cli.BeginLoading("Installing dependencies...")

	_, err = exec.Execute(true, "go", "mod", "tidy")
	if err != nil {
		return err
	}

	cli.FinishLoading()
	return nil
}

func createServerTemplate(projectName string) error {
	return nil
}

func executeServerTemplate(templateText, fileName, projectName, libraryURL, modulePath string) error {
	type data struct {
	}

	return ExecTemplate(templateText, fileName, data{})
}
