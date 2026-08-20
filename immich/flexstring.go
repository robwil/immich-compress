package immich

import (
	"encoding/json"
	"fmt"
)

// FlexString unmarshals both JSON strings and numbers into a Go string.
// This handles API fields where the server may return either type.
type FlexString string

func (f *FlexString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexString(s)
		return nil
	}

	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexString(n.String())
		return nil
	}

	return fmt.Errorf("FlexString: cannot unmarshal %s", string(data))
}

func (f FlexString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

func (f FlexString) String() string {
	return string(f)
}
