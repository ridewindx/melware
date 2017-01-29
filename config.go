package melware

import (
	"github.com/spf13/viper"
	"github.com/ridewindx/mel"
)

// Key for storing config object into a Mel instance.
const ConfigKey = "Config"

type Config struct {
	*viper.Viper
}

// GetConfig returns the config for a Mel instance.
// If the config doesn't exist in Mel, create a new one
// and put it into Mel.
// The config object inherently has all the methods of viper.Viper.
func GetConfig(mel *mel.Mel) *Config {
	v, ok := mel.GetVar(ConfigKey)
	if ok {
		return v.(*Config)
	}

	c := &Config{
		Viper: viper.New(),
	}

	c.ReadInConfig()

	mel.SetVar(ConfigKey, c)
	return c
}

// ReadFile reads config from the specified file.
func (c *Config) ReadFile(configFile string) error {
	c.SetConfigFile(configFile)
	return c.ReadInConfig()
}

// ReadDirs reads config from a found file named as configName in several directories.
func (c *Config) ReadDirs(configName string, dirs ...string) error {
	c.SetConfigName(configName)
	for _, dir := range dirs {
		c.AddConfigPath(dir)
	}
	return c.ReadInConfig()
}
