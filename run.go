package main

import (
	"fmt"
	"net"
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
	runtime, _ := config.LangConfig["runtime"].(string)
	if runtime != "node" && runtime != "bundler" && (typescript || runtime != "browser") {
		return fmt.Errorf("Invalid runtime: '%s'", runtime)
	}

	url := external.TrimURL(config.URL)

	switch config.Type {
	case "client":
		return runClient(url, data.Args, runtime, typescript)
	case "server":
		return runServer(data.Args, typescript)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func runClient(url string, args []string, runtime string, typescript bool) error {
	if typescript && runtime != "bundler" {
		_, err := cgExec.Execute(true, "npx", "tsc")
		if err != nil {
			return err
		}
	}

	if runtime == "node" {
		path := "src/index.js"
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
		var cmdArgs []string
		if runtime == "bundler" {
			cmdArgs = []string{"parcel", "--watch-for-stdin", "-p", "5000", "src/index.html"}
		} else {
			cmdArgs = []string{"serve", "-n", "--no-port-switching", "-p", "5000", "."}
		}

		done := make(chan struct{})
		runWhenUP("localhost:5000", func() {
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

			cgExec.OpenBrowser(fmt.Sprintf("http://localhost:5000?game_url=%s&op=%s%s", url, op, queryParams))
		}, done)

		_, err := cgExec.Execute(false, "npx", cmdArgs...)
		close(done)
		if err != nil {
			return fmt.Errorf("Failed to run '%s': %s", strings.Join(cmdArgs, " "), err)
		}
	}

	return nil
}

func runWhenUP(address string, fn func(), done <-chan struct{}) {
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			con, err := net.Dial("tcp", address)
			if err == nil {
				con.Close()
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		fn()
	}()
}

func runServer(args []string, typescript bool) error {
	panic("not implemented")
	return nil
}
