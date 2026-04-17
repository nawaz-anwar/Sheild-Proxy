package config

import (
"os"

"gopkg.in/yaml.v3"
)

type Config struct {
Phase string `yaml:"phase"`
Proxy struct {
VerifiedHeader string `yaml:"verified_header"`
ClientIDHeader string `yaml:"client_id_header"`
Domains        []struct {
Host     string `yaml:"host"`
Upstream string `yaml:"upstream"`
ClientID string `yaml:"client_id"`
Active   bool   `yaml:"active"`
} `yaml:"domains"`
} `yaml:"proxy"`
}

func Load(path string) (*Config, error) {
b, err := os.ReadFile(path)
if err != nil {
return nil, err
}
var cfg Config
if err := yaml.Unmarshal(b, &cfg); err != nil {
return nil, err
}
return &cfg, nil
}
