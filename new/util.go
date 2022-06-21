package new

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/codegame-cli/util"
)

func NPMVersion(owner, pkg, version string) (string, error) {
	res, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s/%s", owner, pkg))
	if err != nil || res.StatusCode != http.StatusOK || !util.HasContentType(res.Header, "application/json") {
		return "", cli.Error("Couldn't access version information from 'https://registry.npmjs.org/%s/%s'.", owner, pkg)
	}
	defer res.Body.Close()
	type response struct {
		Versions map[string]any `json:"versions"`
	}
	var data response
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil || data.Versions == nil {
		return "", cli.Error("Couldn't decode npm version data.")
	}

	versions := make([]string, 0, len(data.Versions))

	for v := range data.Versions {
		if strings.HasPrefix(v, version) {
			versions = append(versions, v)
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		a := versions[i]
		b := versions[j]

		a1, a2, a3, err := util.ParseVersion(a)
		if err != nil {
			return false
		}
		b1, b2, b3, err := util.ParseVersion(b)
		if err != nil {
			return false
		}

		return a1 > b1 || (a1 == b1 && a2 > b2) || (a1 == b1 && a2 == b2 && a3 > b3)
	})

	if len(versions) > 0 {
		return versions[0], nil
	}

	return "", cli.Error("Couldn't fetch the correct library package version to use.")
}
