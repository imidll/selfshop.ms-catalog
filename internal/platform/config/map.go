package config

import "time"

type T struct {
	App   App   `koanf:"app"   validate:"required"`
	Log   Log   `koanf:"log"   validate:"required"`
	Entry Entry `koanf:"entry" validate:"required"`
	Debug bool  `koanf:"debug"`
}

type App struct {
	Name    string `koanf:"name"    validate:"required"`
	Runmode string `koanf:"runmode" validate:"required,oneof=dev prod"`
}

type Log struct {
	MinLevel string `koanf:"min_level" validate:"required,oneof=debug info warn error panic fatal"`
	Format   string `koanf:"format"    validate:"required,oneof=json console auto"`
}

type Entry struct {
	HTTP HTTP `koanf:"http" validate:"required"`
}

type HTTP struct {
	Port           uint16        `koanf:"port"            validate:"required,min=1024,max=65535"`
	WriteTimeout   time.Duration `koanf:"write_timeout"   validate:"required,gte=5s,lte=90s,gtefield=ReadTimeout"`
	ReadTimeout    time.Duration `koanf:"read_timeout"    validate:"required,gte=5s,lte=60s"`
	IdleTimeout    time.Duration `koanf:"idle_timeout"    validate:"required,gte=30s,lte=180s,gtefield=RequestTimeout"`
	RequestTimeout time.Duration `koanf:"request_timeout" validate:"required,gte=10s,lte=120s"`
}
