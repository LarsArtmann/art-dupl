package filter

import (
	"github.com/LarsArtmann/gogenfilter"
)

type SQLCConfig = gogenfilter.SQLCConfig
type SQLCEngine = gogenfilter.SQLCEngine
type SQLCGenConfig = gogenfilter.SQLCGenConfig
type SQLCGoConfig = gogenfilter.SQLCGoConfig

func FindSQLCConfigs(paths []string) (map[string]string, error) {
	return gogenfilter.FindSQLCConfigs(paths)
}

func ParseSQLCConfig(configPath string) (*SQLCConfig, error) {
	return gogenfilter.ParseSQLCConfig(configPath)
}

func GetSQLOutputDirs(paths []string) ([]string, error) {
	return gogenfilter.GetSQLOutputDirs(paths)
}
