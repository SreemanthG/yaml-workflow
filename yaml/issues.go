package yaml

import "github.com/lyraproj/issue/issue"

const NotYamlStep = `YAML_NOT_STEP`

func init() {
	issue.Hard(NotYamlStep, `a step must contain one of the keys 'action', 'call', 'resource', or 'steps'`)
}
