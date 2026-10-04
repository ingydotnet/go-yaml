# TOML parser plugin for go-yaml

This optional Go module adapts
[`yamlstar-plugin-parser-toml`](https://github.com/yamlstar/yamlstar-plugin-parser-toml)
to `yaml.ParserPlugin`.

```go
import toml "go.yaml.in/yaml/v4/plugin/parser/toml"

err := yaml.Load(input, &value, yaml.WithPlugin(toml.New()))
```

Call `toml.Register()` before using `yaml.OptsYAML` with
`plugin: {parser: toml@v0.1.0}`.
