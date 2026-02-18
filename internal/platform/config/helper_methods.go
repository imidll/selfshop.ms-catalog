package config

func (conf *T) IsDevelopment() bool {
	return conf.App.Runmode == AppRunmodeDev
}

func (conf *T) IsProduction() bool {
	return conf.App.Runmode == AppRunmodeProd
}

func (conf *T) GetLogFormat() string {
	if fmt := conf.Log.Format; fmt != LogFormatAuto {
		return fmt
	}
	if conf.IsProduction() {
		return LogFormatJSON
	}
	return LogFormatConsole
}
