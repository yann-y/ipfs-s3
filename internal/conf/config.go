package conf

type Config struct {
	Http Http `yaml:"http"`
	Db   Db   `yaml:"db"`
}

type Db struct {
	Mysql Mysql `yaml:"mysql"`
}

type Mysql struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Http struct {
	Server Server `yaml:"server"`
}

type Server struct {
	Port int `yaml:"port"`
}
