package cfg

import (
	"github.com/spf13/viper"
)

type CH struct {
	Interval			int64	`mapstructure:"max-interval"`
	Count				int64	`mapstructure:"max-count"`
	ConnectionString	string	`mapstructure:"connection-string"`
}

type Listener struct {
	IP		string		`mapstructure:"listen-ip"`
	Port	int64		`mapstructure:"listen-port"`
	Workers	int			`mapstructure:"workers"`
}

type Cfg struct {
	Ch			CH				`mapstructure:"ch"`
	Listener	Listener		`mapstructure:"listener"`
}

func NewConfig(path string) (*Cfg, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Cfg{}

	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
