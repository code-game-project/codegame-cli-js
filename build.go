package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/modules"
	cp "github.com/otiai10/copy"
)

func Build() error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	data, err := modules.ReadCommandConfig[modules.BuildData]()
	if err != nil {
		return err
	}
	if data.Output == "" {
		data.Output = "build"
	}

	typescript := data.Lang == "ts"

	runtime, _ := config.LangConfig["runtime"].(string)
	if runtime != "node" && runtime != "bundler" && (typescript || runtime != "browser") {
		return fmt.Errorf("Invalid runtime: '%s'", runtime)
	}

	switch config.Type {
	case "client":
		return buildClient(config.Game, data.Output, config.URL, typescript, runtime)
	case "server":
		return buildServer()
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func buildClient(gameName, output, url string, typescript bool, runtime string) error {
	yes, err := cli.YesNo(fmt.Sprintf("The '%s' directory will be completely overwritten. Continue?", output), false)
	if err != nil || !yes {
		return cli.ErrCanceled
	}
	os.RemoveAll(output)
	err = os.MkdirAll(output, 0o755)
	if err != nil {
		return fmt.Errorf("Failed to create output directory: %w", err)
	}

	cli.BeginLoading("Building...")

	if runtime == "node" {
		if typescript {
			_, err = exec.Execute(true, "npx", "tsc", "--outDir", output)
			if err != nil {
				return err
			}
		} else {
			err = cp.Copy("src", output, cp.Options{
				OnSymlink: func(src string) cp.SymlinkAction {
					return cp.Deep
				},
			})
			if err != nil {
				return fmt.Errorf("Failed to copy source files to output directory: %s", err)
			}
		}
	} else if runtime == "bundler" {
		gameJSPath := filepath.Join("src", gameName, "game.js")
		if typescript {
			gameJSPath = filepath.Join("src", gameName, "game.ts")
		}
		err = replaceInFile(gameJSPath, "throw 'Query parameter \"game_url\" must be set.'", fmt.Sprintf("return '%s'", url))
		if err != nil {
			return err
		}

		_, err = exec.Execute(true, "npx", "parcel", "build", "--dist-dir", output, "src/index.html")
		if err != nil {
			return err
		}

		err = replaceInFile(gameJSPath, fmt.Sprintf("return '%s'", url), "throw 'Query parameter \"game_url\" must be set.'")
		if err != nil {
			return err
		}
	} else if runtime == "browser" {
		err = cp.Copy(".", output, cp.Options{
			Skip: func(src string) (bool, error) {
				return src == filepath.Clean(output) || (src != "node_modules" && strings.HasPrefix(src, "node_modules") && !strings.Contains(src, "@code-game-project")) || src == ".codegame.json" || src == "package.json" || src == "package-lock.json", nil
			},
		})
		if err != nil {
			return fmt.Errorf("Failed to copy source files to output directory: %s", err)
		}
	}

	if runtime != "bundler" {
		gameJSPath := filepath.Join(output, gameName, "game.js")
		if runtime == "node" {
			err = replaceInFile(gameJSPath, "throw 'Environment variable `CG_GAME_URL` must be set.'", fmt.Sprintf("return '%s'", url))
		} else {
			err = replaceInFile(gameJSPath, "throw 'Query parameter \"game_url\" must be set.'", fmt.Sprintf("return '%s'", url))
		}
	}
	if err != nil {
		return err
	}
	cli.FinishLoading()
	return nil
}

func replaceInFile(filename, old, new string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Failed to replace '%s' with '%s' in '%s': %s", old, new, filename, err)
	}
	content = []byte(strings.ReplaceAll(string(content), old, new))
	err = os.WriteFile(filename, content, 0o644)
	if err != nil {
		return fmt.Errorf("Failed to replace '%s' with '%s' in '%s': %s", old, new, filename, err)
	}
	return nil
}

func buildServer() error {
	panic("not implemented")
	return nil
}
