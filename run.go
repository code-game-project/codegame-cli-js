package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/code-game-project/go-utils/cgfile"
	cgExec "github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/external"
	"github.com/code-game-project/go-utils/modules"
)

func Run() error {
	data, err := modules.ReadCommandConfig[modules.RunData]()
	if err != nil {
		return err
	}

	typescript := data.Lang == "ts"

	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}
	runtime := config.LangConfig["runtime"]
	node := runtime == "node"

	if !typescript || !node {
		panic("not implemented")
	}

	url := external.TrimURL(config.URL)

	switch config.Type {
	case "client":
		return runClient(url, data.Args, typescript, node)
	case "server":
		return runServer(data.Args, typescript, node)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func runClient(url string, args []string, typescript, node bool) error {
	if typescript {
		_, err := cgExec.Execute(true, "npx", "tsc")
		if err != nil {
			return err
		}
	}

	cmdArgs := []string{"dist/index.js"}
	cmdArgs = append(cmdArgs, args...)

	env := []string{"CG_GAME_URL=" + url}
	env = append(env, os.Environ()...)

	if _, err := exec.LookPath("node"); err != nil {
		return fmt.Errorf("'node' ist not installed!")
	}

	cmd := exec.Command("node", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run 'CG_GAME_URL=%s node %s'", url, strings.Join(cmdArgs, " "))
	}
	return nil
}

func runServer(args []string, typescript, node bool) error {
	panic("not implemented")
	return nil
}
