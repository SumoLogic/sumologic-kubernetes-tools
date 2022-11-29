package falcoupgrade

type Values struct {
	Falco struct {
		Falco struct {
			JsonOutputOld *bool                  `yaml:"jsonOutput,omitempty"`
			JsonOutputNew *bool                  `yaml:"json_output,omitempty"`
			RulesFileOld  *[]interface{}         `yaml:"rulesFile,omitempty"`
			RulesFileNew  *[]interface{}         `yaml:"rules_file,omitempty"`
			LoadPlugins   *[]string              `yaml:"load_plugins,omitempty"`
			Rest          map[string]interface{} `yaml:",inline"`
		} `yaml:"falco,omitempty"`
		Image               map[string]interface{} `yaml:"image,omitempty"`
		Enabled             *bool                  `yaml:"enabled,omitempty"`
		ExtraInitContainers *[]interface{}         `yaml:"extraInitContainers,omitempty"`
		Extra               struct {
			InitContainers *[]interface{} `yaml:"initContainers,omitempty"`
		} `yaml:"extra,omitempty"`
		Ebpf *struct {
			Enabled *bool                  `yaml:"enabled,omitempty"`
			Rest    map[string]interface{} `yaml:",inline"`
		} `yaml:"ebpf,omitempty"`
		Driver *struct {
			Kind *string `yaml:"kind,omitempty"`
		} `yaml:"driver,omitempty"`
		Rest map[string]interface{} `yaml:",inline"`
	} `yaml:"falco,omitempty"`
	Rest map[string]interface{} `yaml:",inline"`
}
