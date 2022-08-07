package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/code-game-project/go-utils/cgfile"
	cgExec "github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/external"
	"github.com/code-game-project/go-utils/modules"
	"github.com/spf13/pflag"
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

	if node {
		path := "index.js"
		if typescript {
			path = "dist/index.js"
		}

		cmdArgs := []string{path}
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
			return fmt.Errorf("Failed to run 'CG_GAME_URL=%s node %s'.", url, strings.Join(cmdArgs, " "))
		}
	} else {
		if _, err := exec.LookPath("npx"); err != nil {
			return fmt.Errorf("'npx' ist not installed!")
		}

		cmd := exec.Command("npx", "serve", "-n", "--no-port-switching", "-l", "5000", ".")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("Failed to start 'npx serve': %s", err)
		}

		pflag.Usage = func() {
			fmt.Fprintf(os.Stdout, "Usage: %s [options] [command]\n\n", os.Args[0])
			fmt.Fprintf(os.Stdout, "Options:\n")
			pflag.PrintDefaults()
			fmt.Fprintf(os.Stdout, "\nCommands:\n")
			fmt.Fprintf(os.Stdout, "  create [options] <username>                        Create and join a new game.\n")
			fmt.Fprintf(os.Stdout, "  join [options] <game_id> <username> [join_secret]  Join an existing game.\n")
			fmt.Fprintf(os.Stdout, "  reconnect [options] <username>                     Reconnect to a game using a saved session.\n")
			fmt.Fprintf(os.Stdout, "  spectate [options] [game_id]                       Spectate a new or existing game.\n")
		}

		var queryParams string
		var public bool
		pflag.BoolVar(&public, "public", false, "Make the created game protected.")
		var protected bool
		pflag.BoolVar(&public, "protected", false, "Make the created game protected.")
		pflag.CommandLine.Parse(args)

		var op string
		if pflag.NArg() > 0 {
			op = pflag.Arg(0)
		}

		switch op {
		case "create", "reconnect":
			if pflag.NArg() > 1 {
				queryParams += "&username=" + pflag.Arg(1)
			}
		case "join":
			if pflag.NArg() > 1 {
				queryParams += "&game_id=" + pflag.Arg(1)
			}
			if pflag.NArg() > 2 {
				queryParams += "&username=" + pflag.Arg(2)
			}
			if pflag.NArg() > 3 {
				queryParams += "&join_secret=" + pflag.Arg(3)
			}
		case "spectate":
			if pflag.NArg() > 1 {
				queryParams += "&game_id=" + pflag.Arg(1)
			}
		}

		if public {
			queryParams += "&public=true"
		}
		if protected {
			queryParams += "&protected=true"
		}

		time.Sleep(2 * time.Second)

		cgExec.OpenBrowser(fmt.Sprintf("http://localhost:5000?game_url=%s&op=%s%s", url, op, queryParams))
		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("Failed to run 'npx serve -n --no-port-switching -l 5000 .': %s", err)
		}
	}

	return nil
}

func runServer(args []string, typescript, node bool) error {
	panic("not implemented")
	return nil
}
