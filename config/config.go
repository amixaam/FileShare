// config/config.go
package config

import (
	"flag"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
    // CLI and YAML configurable
    Directory string `yaml:"directory"`
    Port      string `yaml:"port"`
    
    // YAML only settings
    ShowDotfiles bool   `yaml:"dotfiles"`  // Controls visibility of hidden files
    Domain       string `yaml:"domain"`    // Domain for links
}

func LoadConfig() (*Config, error) {
    // Default configuration
    cfg := &Config{
        Port:        "8080",
        ShowDotfiles: true,     // Default to showing hidden files
        Domain:      "localhost", // Default domain
    }

    // Get home directory for default shared files location
    homeDir, err := os.UserHomeDir()
    if err == nil {
        cfg.Directory = filepath.Join(homeDir, "SharedFiles")
    }

    // Define CLI flags (only for directory and port)
    configFile := flag.String("config", "fileshare.yaml", "Path to config file")
    flag.StringVar(&cfg.Directory, "dir", cfg.Directory, "Directory to serve")
    flag.StringVar(&cfg.Port, "port", cfg.Port, "Port to listen on")
    flag.Parse()

    // Try to load config file
    if err := loadConfigFile(*configFile, cfg); err != nil {
        // Only return error if the config file exists but couldn't be parsed
        if !os.IsNotExist(err) {
            return nil, err
        }
    }

    // CLI flags override config file settings (only for directory and port)
    flag.Visit(func(f *flag.Flag) {
        switch f.Name {
        case "dir":
            cfg.Directory = f.Value.String()
        case "port":
            cfg.Directory = f.Value.String()
        }
    })

    // Resolve absolute path for directory
    if cfg.Directory != "" {
        absPath, err := filepath.Abs(cfg.Directory)
        if err == nil {
            cfg.Directory = absPath
        }
    }

    return cfg, nil
}

func loadConfigFile(path string, cfg *Config) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    decoder := yaml.NewDecoder(file)
    return decoder.Decode(cfg)
}