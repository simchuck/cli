package config

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func eq(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func Test_parseConfig(t *testing.T) {
	defer StubConfig(`---
hosts:
  github.com:
    user: monalisa
    oauth_token: OTOKEN
`, "")()
	config, err := ParseConfig("config.yml")
	eq(t, err, nil)
	user, err := config.Get("github.com", "user")
	eq(t, err, nil)
	eq(t, user, "monalisa")
	token, err := config.Get("github.com", "oauth_token")
	eq(t, err, nil)
	eq(t, token, "OTOKEN")
}

func Test_parseConfig_multipleHosts(t *testing.T) {
	defer StubConfig(`---
hosts:
  example.com:
    user: wronguser
    oauth_token: NOTTHIS
  github.com:
    user: monalisa
    oauth_token: OTOKEN
`, "")()
	config, err := ParseConfig("config.yml")
	eq(t, err, nil)
	user, err := config.Get("github.com", "user")
	eq(t, err, nil)
	eq(t, user, "monalisa")
	token, err := config.Get("github.com", "oauth_token")
	eq(t, err, nil)
	eq(t, token, "OTOKEN")
}

func Test_parseConfig_hostsFile(t *testing.T) {
	defer StubConfig("", `---
github.com:
  user: monalisa
  oauth_token: OTOKEN
`)()
	config, err := ParseConfig("config.yml")
	eq(t, err, nil)
	user, err := config.Get("github.com", "user")
	eq(t, err, nil)
	eq(t, user, "monalisa")
	token, err := config.Get("github.com", "oauth_token")
	eq(t, err, nil)
	eq(t, token, "OTOKEN")
}

func Test_parseConfig_notFound(t *testing.T) {
	defer StubConfig(`---
hosts:
  example.com:
    user: wronguser
    oauth_token: NOTTHIS
`, "")()
	config, err := ParseConfig("config.yml")
	eq(t, err, nil)
	_, err = config.Get("github.com", "user")
	eq(t, err, &NotFoundError{errors.New(`could not find config entry for "github.com"`)})
}

func Test_ParseConfig_migrateConfig(t *testing.T) {
	defer StubConfig(`---
github.com:
  - user: keiyuri
    oauth_token: 123456
`, "")()

	mainBuf := bytes.Buffer{}
	hostsBuf := bytes.Buffer{}
	defer StubWriteConfig(&mainBuf, &hostsBuf)()
	defer StubBackupConfig()()

	_, err := ParseConfig("config.yml")
	eq(t, err, nil)

	expectedMain := ""
	expectedHosts := `github.com:
    user: keiyuri
    oauth_token: "123456"
`

	eq(t, mainBuf.String(), expectedMain)
	eq(t, hostsBuf.String(), expectedHosts)
}

func Test_parseConfigFile(t *testing.T) {
	fileContents := []string{"", " ", "\n"}
	for _, contents := range fileContents {
		t.Run(fmt.Sprintf("contents: %q", contents), func(t *testing.T) {
			defer StubConfig(contents, "")()
			_, yamlRoot, err := parseConfigFile("config.yml")
			eq(t, err, nil)
			eq(t, yamlRoot.Content[0].Kind, yaml.MappingNode)
			eq(t, len(yamlRoot.Content[0].Content), 0)
		})
	}
}
