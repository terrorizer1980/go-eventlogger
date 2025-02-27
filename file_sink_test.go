package eventlogger

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestFileSink_NewDir(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sinkDir := filepath.Join(tmpDir, "file_sink")

	fs := FileSink{
		Path:     sinkDir,
		FileName: "audit.log",
	}

	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first")},
		Payload:   "First entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"audit.log"}
	files, _ := ioutil.ReadDir(sinkDir)
	got := []string{}
	for _, f := range files {
		got = append(got, f.Name())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected %v files, got %v file(s)", want, got)
	}
}

func TestFileSink_Reopen(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs := FileSink{
		Path:     tmpDir,
		FileName: "audit.log",
	}
	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first")},
		Payload:   "First entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	// delete file
	err = os.Remove(filepath.Join(tmpDir, "audit.log"))
	if err != nil {
		t.Fatal(err)
	}

	// reopen
	err = fs.Reopen()
	if err != nil {
		t.Fatal(err)
	}

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("second")},
		Payload:   "Second entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure process re-created the file
	dat, err := ioutil.ReadFile(filepath.Join(tmpDir, "audit.log"))
	if err != nil {
		t.Fatal(err)
	}

	got := string(dat)
	want := "second"
	if got != "second" {
		t.Errorf("Expected file content to be %s, got %s", want, got)
	}

	files := 1
	if got, _ := ioutil.ReadDir(tmpDir); len(got) != files {
		t.Errorf("Expected %d files, got %v file(s)", files, len(got))
	}
}

func TestFileSink_TimeRotate(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxDuration: 2 * time.Second,
	}
	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first")},
		Payload:   "First entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first")},
		Payload:   "First entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	want := 2
	if got, _ := ioutil.ReadDir(tmpDir); len(got) != want {
		t.Errorf("Expected %d files, got %v file(s)", want, len(got))
	}
}

func TestFileSink_ByteRotate(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxBytes:    5,
		MaxDuration: 24 * time.Hour,
	}
	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("entry")},
		Payload:   "entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("entry")},
		Payload:   "entry",
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	want := 2
	if got, _ := ioutil.ReadDir(tmpDir); len(got) != want {
		t.Errorf("Expected %d files, got %v file(s)", want, len(got))
	}
}

func TestFileSink_open(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxDuration: 1 * time.Second,
	}
	err = fs.open()
	if err != nil {
		t.Fatal(err)
	}

	_, err = ioutil.ReadFile(fs.f.Name())
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileSink_pruneFiles(t *testing.T) {
	t.Parallel()

	tmpDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fs := FileSink{
		Path:        tmpDir,
		FileName:    "audit.log",
		MaxDuration: 1 * time.Hour,
		MaxBytes:    10,
		MaxFiles:    1,
	}

	event := &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("first entry")},
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("second entry")},
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	event = &Event{
		Formatted: map[string][]byte{JSONFormat: []byte("third entry")},
	}
	_, err = fs.Process(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}

	want := 2
	tmpFiles, _ := ioutil.ReadDir(tmpDir)
	got := len(tmpFiles)
	if want != got {
		t.Errorf("Expected %d files, got %d", want, got)
	}
}
