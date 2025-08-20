package cresettings

import (
	_ "embed"
	"encoding/json"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update the golden files of this test")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

//go:generate go test . -update
var (
	//go:embed defaults.json
	defaultsJSON string
	//go:embed defaults.toml
	defaultsTOML string
)

func TestDefault(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		b, err := json.MarshalIndent(Default, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
		if *update {
			require.NoError(t, os.WriteFile("defaults.json", b, 0644))
		} else {
			require.Equal(t, defaultsJSON, string(b))
		}
	})

	t.Run("toml", func(t *testing.T) {
		b, err := toml.Marshal(Default)
		if err != nil {
			log.Fatal(err)
		}
		if *update {
			require.NoError(t, os.WriteFile("defaults.toml", b, 0644))
		} else {
			require.Equal(t, defaultsTOML, string(b))
		}
	})
}
