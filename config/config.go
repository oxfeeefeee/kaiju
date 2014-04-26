package config

import (
    "os"
    "encoding/json"
    "github.com/oxfeeefeee/kaiju/log"
)

type Config struct {
    SeedPeers []string
}

var cfg *Config

func GetConfig() *Config {
    return cfg
}

func ReadJsonConfigFile() error {
    wd, wderr := os.Getwd()
    if wderr == nil {
        log.MainLogger().Printf("Working directory: %s", wd)
    } else {
        return wderr
    }

    configFile, err := os.Open("./config.json")
    if err != nil {
        return err
    }
    cfg = new(Config)
    jsonParser := json.NewDecoder(configFile)
    if err = jsonParser.Decode(cfg); err != nil {
        return err
    }
    return nil
}