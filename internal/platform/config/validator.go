package config

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

func validate(conf *T) error {
	v := validator.New(validator.WithRequiredStructEnabled())

	if err := v.Struct(conf); err != nil {
		return err
	}

	var errs []error

	if conf.IsProduction() {
		if conf.Debug {
			errs = append(errs, errors.New("debug must be false in prod mode"))
		}
		if conf.Log.MinLevel == LogMinLevelDebug {
			errs = append(errs, errors.New("log min level must be info or higher in prod"))
		}
		if conf.Log.Format == LogFormatConsole {
			errs = append(errs, errors.New("log format must be json in prod"))
		}
	}

	return errors.Join(errs...)
}
