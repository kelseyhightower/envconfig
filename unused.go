package envconfig

import (
	"os"
	"regexp"
	"strings"
)

var envVarNameRegex = regexp.MustCompile("^[^=]*")

// Unused returns the list of env vars that are defined right now
// but are not being parsed by envconfig for the prefix given
// If prefix is empty, it will check all the env vars available
// and return the ones that are not being parsed.
func Unused(prefix string, spec interface{}) ([]string, error) {
	re, err := applicableRegex(prefix)
	if err != nil {
		return nil, err
	}

	// build a map with used env vars as keys
	infos, err := gatherInfo(prefix, spec)
	used := make(map[string]struct{}, len(infos))
	for _, v := range infos {
		used[v.Key] = struct{}{}
		if v.Alt != "" {
			used[v.Alt] = struct{}{}
		}
	}

	var unused []string

	// read all the defined env vars, if begins with same prefix and is not in the previous map, it's unused
	for _, envName := range envVarsNames() {
		if !re.MatchString(envName) {
			continue
		}
		if _, used := used[envName]; envName != "" && !used {
			unused = append(unused, envName)
		}
	}

	return unused, nil
}

func envVarsNames() []string {
	environ := os.Environ()
	names := make([]string, len(environ))
	for i, env := range environ {
		names[i] = envVarNameRegex.FindString(env)
	}
	return names
}

func applicableRegex(prefix string) (*regexp.Regexp, error) {
	if prefix == "" {
		return regexp.Compile(".+")
	}
	upperPrefix := strings.ToUpper(prefix)
	return regexp.Compile("^" + upperPrefix + "_")
}
