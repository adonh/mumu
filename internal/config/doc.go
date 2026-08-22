// Package config resolves and loads mumu's user-editable configuration
// file, an explicit YAML file that controls where mumu stores its data.
//
// Convention: any YAML mumu writes uses two-space indentation. config.yaml
// is currently a flat, hand-written string (see defaultConfigYAML), so this
// has no visible effect yet, but if a future setting needs a YAML encoder
// (nested structure, lists), configure its indent width to 2
// (e.g. yaml.Encoder.SetIndent(2), or the equivalent for whichever YAML
// library is in use) rather than accepting the library's default.
package config
