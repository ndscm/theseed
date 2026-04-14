package dotenv

import (
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func ReadEnvFile(filePath string) (map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	result, err := Parse(string(data))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return result, nil
}

func Load(filePaths ...string) error {
	merged := make(map[string]string)
	for _, fp := range filePaths {
		envMap, err := ReadEnvFile(fp)
		if err != nil {
			return err
		}
		for key, value := range envMap {
			merged[key] = value
		}
	}

	for key, value := range merged {
		err := os.Setenv(key, value)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

func LoadAncestor(workDir string, fileName string) error {
	dir, err := filepath.Abs(workDir)
	if err != nil {
		return seederr.Wrap(err)
	}

	candidates := []string{}
	for {
		candidate := filepath.Join(dir, fileName)
		_, err = os.Stat(candidate)
		if err == nil {
			candidates = append([]string{candidate}, candidates...)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return Load(candidates...)
}
