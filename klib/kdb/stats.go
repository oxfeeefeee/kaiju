package kdb

import (
    "fmt"
)

type Stats struct {
    capacity        uint32
    // valid records
    records         uint32
    // Slots that are occupied by deleted keys
    deadSlots       uint32
    // Values marked as deleted
    deadValues      uint32
}

func (s *Stats) String() string {
    f := "KDB Stats:[capacity:%d, records:%d deadSlots:%d, deadValues:%d]"
    return fmt.Sprintf(f, s.capacity, s.records, s.records, s.deadValues)      
}

func (s *Stats) Saturation() float32 {
    return float32((s.records + s.deadSlots)  * 10000 / s.capacity) / 100.0
}

func (s *Stats) Capacity() uint32 {
    return s.capacity
}

func (s *Stats) Records() uint32 {
    return s.records
}

func (s *Stats) DeadSlots() uint32 {
    return s.deadSlots
}

func (s *Stats) DeadValues() uint32 {
    return s.deadValues
}
