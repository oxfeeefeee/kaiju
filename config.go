package kaiju

import (
    "os"
    "errors"
    "path/filepath"
    "encoding/json"
)

const configFileName = "config.json"

type Config struct {
    SeedPeers       []string
    DataDir         string
    TempDataDir     string
    LogFileName     string
}

var cfg             *Config

var configFileDir   string

func GetConfig() *Config {
    if cfg == nil {
        panic("Haven't read config file yet!")
    }
    return cfg
}

func GetConfigFileDir() string {
    return configFileDir
}

func ReadJsonConfigFile() error {
    if cfg != nil {
        return nil
    }

    // Search up for config.json
    left, upALevel, right, result := "./", "../", configFileName, ""
    for i := 0; i < 5; i ++ {
        p := filepath.Join(left, right)
        if _, err := os.Stat(p); err == nil {
            configFileDir = left
            result = p
            break
        }
        left = filepath.Join(left, upALevel)
    }

    if result == "" {
        return errors.New("Couldn't find config file")
    }

    configFile, err := os.Open(result)
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