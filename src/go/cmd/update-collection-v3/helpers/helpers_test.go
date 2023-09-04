package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForDuplicatedValue(t *testing.T) {
	type nest struct {
		str  string                 `yaml:"str,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	}

	type noRest struct {
		str string `yaml:"str,omitempty"`
	}

	type main struct {
		Str    string                 `yaml:"str,omitempty"`
		i      int                    `yaml:"int,omitempty"`
		Nest   nest                   `yaml:"nest,omitempty"`
		Ptr    *nest                  `yaml:"ptr,omitempty"`
		Rest   map[string]interface{} `yaml:",inline"`
		NoRest noRest                 `yaml:"noRest,omitempty"`
	}

	a := &main{
		Rest: map[string]interface{}{
			"str": "",
		},
		Nest: nest{
			str: "test",
			Rest: map[string]interface{}{
				"str": "test",
			},
		},
		Ptr: &nest{
			Rest: map[string]interface{}{
				"str": "test",
			},
		},
		i: 3,
	}

	ret, err := CheckForConflictsInRest(a)
	assert.EqualError(t, err, "conflict between input and output values for the following keys: str, nest.str, ptr.str")
	assert.Equal(t, []string{"str", "nest.str", "ptr.str"}, ret)
}
