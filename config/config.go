package config

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

// Config defines all program settigs
type Config struct {
	ElectionName string
	DatabaseFile string
	StartTime    time.Time
	EndTime      time.Time
	DiscordKey   string
}

// LoadConfig creates a new config from a file
func LoadConfig(filename string) (Config, error) {
	var conf Config
	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		return conf, fmt.Errorf("Unable to load config %s: %v", filename, err)
	}
	return conf, nil
}

func main() {
	fmt.Printf("Pi: %f\n", 3.1235235)
}
