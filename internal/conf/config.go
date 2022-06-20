package conf

type Config struct {
	Http    Http    `yaml:"http"`
	Db      Db      `yaml:"db"`
	Storage Storage `yaml:"storage"`
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

type Storage struct {
	Model string `yaml:"model"`
	Host  string `yaml:"host"`
}
