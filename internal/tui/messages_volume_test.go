package tui

import (
    "fmt"
    "testing"
)

func TestHighVolumeConsistency(t *testing.T) {
    m := initialModel()

    total := 500
    for i := 0; i < total; i++ {
        if i%2 == 0 {
            m.textarea.SetValue(fmt.Sprintf("msg-%d", i))
        } else {
            m.textarea.SetValue("add(1, 2)")
        }
        m.handleEnterKey()
        if len(m.messages) != len(m.storageMessages) {
            t.Fatalf("divergence at %d: display %d storage %d", i, len(m.messages), len(m.storageMessages))
        }
    }
}