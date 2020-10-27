package patrol

import (
	"encoding/json"
	"time"
)

type duration time.Duration

func (d duration) isZero() bool {
	return int64(d) == 0
}

func (d *duration) UnmarshalJSON(buffer []byte) error {
	var str string
	if err := json.Unmarshal(buffer, &str); err != nil {
		return err
	}
	parsedD, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*d = duration(parsedD)
	return nil
}

func (d duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.duration().String() + `"`), nil
}

func (d duration) duration() time.Duration {
	return time.Duration(d)
}
