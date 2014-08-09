package kdb

import (
    "sync"
)

type Stats struct {
    // How many slots are occupied, including slots that are marked as deleted
    occupiedSlotCount   int64
    // How many entries in this DB recordCount = occupiedSlotCount - slots_occupied_by_deleted_items
    recordCount         int64
    // How many scan operations did in total
    scanCount           int64
    // How many slot-read did
    slotReadCount       int64
    // Thread safety
    mutex               sync.RWMutex
}

func (s *Stats) OccupiedSlotCount() int64 {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.occupiedSlotCount
}

func (s *Stats) RecordCount() int64 {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.recordCount
}

func (s *Stats) ScanCount() int64 {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.scanCount
}

func (s *Stats) SlotReadCount() int64 {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.slotReadCount
}

func (s *Stats) incOccupiedSlotCount() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.occupiedSlotCount++
}

func (s *Stats) incRecordCount() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.recordCount++
}

func (s *Stats) incScanCount() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.scanCount++
}

func (s *Stats) incSlotReadCount() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.slotReadCount++
}