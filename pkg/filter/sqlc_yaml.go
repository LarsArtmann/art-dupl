package filter

import (
	"github.com/LarsArtmann/gogenfilter"
)

func FindSQLCConfigs(paths []string) (map[string]string, error) {
	configs, err := gogenfilter.FindSQLCConfigs(paths)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func GetSQLOutputDirs(paths []string) ([]string, error) {
	dirs, err := gogenfilter.GetSQLOutputDirs(paths)
	if err != nil {
		return nil, err
	}

	return dirs, nil
}
