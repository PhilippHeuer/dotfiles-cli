package config

type DotfilesConfig struct {
	Directories []Dir `yaml:"directories"`
}

type Dir struct {
	Path   string `yaml:"path"`
	Target string `yaml:"target"`
}
