package new

import (
	"encoding/json"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Bananenpro/cli"
)

func ExecTemplate(templateText, path string, data any) error {
	err := os.MkdirAll(filepath.Join(filepath.Dir(path)), 0755)
	if err != nil {
		return err
	}

	tmpl, err := template.New(path).Parse(templateText)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(path))
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

func ConfigurePackageJSON() error {
	file, err := os.Open("package.json")
	if err != nil {
		return cli.Error("Failed to open 'package.json'.")
	}

	var data map[string]any
	err = json.NewDecoder(file).Decode(&data)
	file.Close()
	if err != nil {
		return cli.Error("Faile to decode 'package.json'.")
	}

	data["type"] = "module"

	file, err = os.Create("package.json")
	if err != nil {
		return cli.Error("Failed to open 'package.json'.")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
