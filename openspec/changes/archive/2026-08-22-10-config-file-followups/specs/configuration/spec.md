## ADDED Requirements

### Requirement: YAML files use two-space indentation

Any YAML file mumu writes SHALL be encoded using two-space indentation. This applies to `config.yaml` and to any other YAML mumu generates now or in the future.

#### Scenario: Generated config file uses two-space indentation

- **WHEN** the system auto-creates `config.yaml` with default settings
- **THEN** the file's YAML content, if and when it contains nested or list structure, uses two-space indentation
