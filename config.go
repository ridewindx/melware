package melware

import (
	"github.com/spf13/viper"
	"github.com/ridewindx/mel"
)

// Key for storing config object into a Mel instance.
const ConfigKey = "Config"

type config struct {
	viper.Viper
}

// Config returns a config object for the Mel instance.
// If the config object doesn't exist in Mel, create a new one
// and put it into Mel.
// The config object inherently has all the methods of viper.Viper.
func Config(mel *mel.Mel) *config {
	v, ok := mel.Get(ConfigKey)
	if ok {
		return v.(*config)
	}

	c := &config{
		Viper: viper.New(),
	}

	c.ReadInConfig()

	mel.Set(ConfigKey, c)
	return c
}

func GetConfig(c *mel.Context) *config {
	return Config(c.Mel)
}

func (c *config) ReadFile(configFile string) error {
	c.SetConfigFile(configFile)
	return c.ReadInConfig()
}

func (c *config) ReadDirs(configName string, dirs ...string) {
	c.SetConfigName(configName)
	for _, dir := range dirs {
		c.AddConfigPath(dir)
	}
	return c.ReadInConfig()
}
