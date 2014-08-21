package kdb

import (
    "os"
    "testing"
    "path/filepath"
    "github.com/oxfeeefeee/kaiju"
)

func openFile(t *testing.T, path string, n string) *os.File {
    path = filepath.Join(path, n)
    exists, _ := fileExists(path)
    if !exists {
        t.Errorf("File doesn't exist: %s", path)
        return nil
    }
    f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
    if err != nil {
        t.Errorf("Failed to open file %s", err)
        return nil
    }
    return f
}

func TestScanKeys(t *testing.T) {
    cfg := kaiju.GetConfig()
    path := filepath.Join(kaiju.ConfigFileDir(), cfg.DataDir)
    f := openFile(t, path, cfg.KdbFileName)
    wa := openFile(t, path, cfg.KdbWAFileName)
    db, err := Load(f, wa)
    if err != nil {
        t.Errorf("Failed load db %s", err)
    }
    r, g, err := db.enumerate(nil)
    if err != nil {
        t.Errorf("Failed to scan keys: %s", err)
    } 
    t.Logf("KDB scanKeys %d %d",r, g)
}

func _TestRebuild(t *testing.T) {
    cfg := kaiju.GetConfig()
    path := filepath.Join(kaiju.ConfigFileDir(), cfg.DataDir)
    f := openFile(t, path, cfg.KdbFileName)
    wa := openFile(t, path, cfg.KdbWAFileName)
    rebf := createFile(t, path, cfg.KdbFileName + ".reb")
    rebwa := createFile(t, path, cfg.KdbWAFileName + ".reb")
    db, err := Load(f, wa)
    if err != nil {
        t.Errorf("Failed to load db %s", err)
    }
    db2, err := db.Rebuild(cfg.KDBCapacity, rebf, rebwa)
    if err != nil {
        t.Errorf("Failed to rebuild db %s", err)
    }

    r, g, err := db2.enumerate(nil)
    if err != nil {
        t.Errorf("Failed to scan keys: %s", err)
    } 
    t.Logf("KDB scanKeys %d %d",r, g)
}

