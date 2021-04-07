package patrol

import (
	"time"
)

type duration time.Duration

func (d duration) isZero() bool {
	return int64(d) == 0
}

func (d *duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	parsedD, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*d = duration(parsedD)
	return nil
}

func (d duration) MarshalYAML() (interface{}, error) {
	return d.duration().String(), nil
}

func (d duration) duration() time.Duration {
	return time.Duration(d)
}
