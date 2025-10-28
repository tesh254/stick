package provider

import (
    "bufio"
    "os"
    "strings"
    "testing"
)

// mockTeaConsumer emulates a Bubble Tea-like component that reads from a string channel.
type mockTeaConsumer struct{
    received []string
}

func (m *mockTeaConsumer) consume(ch <-chan string) {
    for s := range ch {
        m.received = append(m.received, s)
    }
}

func TestRunCommand_ChannelOutput(t *testing.T) {
    final, ch := runCommand(RunCommandArgs{Command: "echo hello"})
    if final == "" {
        t.Errorf("expected non-empty final context string")
    }
    consumer := &mockTeaConsumer{}
    consumer.consume(ch)
    if len(consumer.received) == 0 {
        t.Fatalf("expected at least one message")
    }
    msg := consumer.received[0]
    if !strings.HasPrefix(msg, "[TOOL_OUTPUT] ") {
        t.Fatalf("message should be prefixed with [TOOL_OUTPUT], got: %s", msg)
    }
    if !strings.Contains(msg, "\"tool\":\"run_command\"") {
        t.Errorf("expected tool run_command in payload")
    }
    if !strings.Contains(msg, "\"status\":\"success\"") {
        t.Errorf("expected success status in payload")
    }
    if !strings.Contains(msg, "hello") {
        t.Errorf("expected raw output 'hello' in payload")
    }
}

func TestReadFile_SuccessAndErrorPropagation(t *testing.T) {
    // success: create a temp file and read it
    path := t.TempDir() + "/file.txt"
    content := "abc123"
    if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
        t.Fatalf("setup failed: %v", err)
    }
    _, ch1 := readFile(ReadFileArgs{Path: path})
    msgs1 := collect(ch1)
    if len(msgs1) == 0 || !strings.Contains(msgs1[0], "\"status\":\"success\"") || !strings.Contains(msgs1[0], content) {
        t.Fatalf("expected success with content, got: %v", msgs1)
    }

    // error: try a missing file
    _, ch2 := readFile(ReadFileArgs{Path: path + ".missing"})
    msgs2 := collect(ch2)
    if len(msgs2) == 0 || !strings.Contains(msgs2[0], "\"status\":\"error\"") {
        t.Fatalf("expected error status, got: %v", msgs2)
    }
}

func TestChannel_ClosesProperly(t *testing.T) {
    _, ch := readDir(ReadDirArgs{Path: "."})
    // Drain until closed; ensure no blocking.
    scanner := bufio.NewScanner(strings.NewReader(strings.Join(collect(ch), "\n")))
    count := 0
    for scanner.Scan() {
        count++
    }
    if count == 0 {
        t.Errorf("expected at least one emitted line")
    }
}

// collect reads entire channel into a slice.
func collect(ch <-chan string) []string {
    var out []string
    for s := range ch {
        out = append(out, s)
    }
    return out
}