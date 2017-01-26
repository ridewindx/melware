package melware

import (
	"github.com/spf13/viper"
	"github.com/ridewindx/mel"
)

// Key for storing config object into a Mel instance.
const ConfigKey = "Config"

type config struct {
	*viper.Viper
}

// Config returns the config for a Mel instance.
// If the config doesn't exist in Mel, create a new one
// and put it into Mel.
// The config object inherently has all the methods of viper.Viper.
func Config(mel *mel.Mel) *config {
	v, ok := mel.GetVar(ConfigKey)
	if ok {
		return v.(*config)
	}

	c := &config{
		Viper: viper.New(),
	}

	c.ReadInConfig()

	mel.SetVar(ConfigKey, c)
	return c
}

// GetConfig gets the Mel config from mel.Context.
func GetConfig(c *mel.Context) *config {
	return Config(c.Mel)
}

// ReadFile reads config from the specified file.
func (c *config) ReadFile(configFile string) error {
	c.SetConfigFile(configFile)
	return c.ReadInConfig()
}

// ReadDirs reads config from a found file named as configName in several directories.
func (c *config) ReadDirs(configName string, dirs ...string) error {
	c.SetConfigName(configName)
	for _, dir := range dirs {
		c.AddConfigPath(dir)
	}
	return c.ReadInConfig()
}
