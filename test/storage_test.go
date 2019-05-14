package test

import (
	"fmt"
	"testing"

	"github.com/scalog/scalog/data/storage"
)

func TestSetupStorage(t *testing.T) {
	_, err := storage.NewStorage("disk0")
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestSingleReadAndWrite(t *testing.T) {
	disk, err := storage.NewStorage("disk1")
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := "Hello, World!"
	writeErr := disk.Write(0, expected)
	if writeErr != nil {
		t.Fatalf(writeErr.Error())
	}
	actual, readErr := disk.Read(0)
	if readErr != nil {
		t.Fatalf(readErr.Error())
	}
	if actual != expected {
		t.Fatalf(fmt.Sprintf("Expected: \"%s\", Actual: %s", expected, actual))
	}
}

func TestMultipleReadAndWrite(t *testing.T) {
	disk, err := storage.NewStorage("disk2")
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected0 := "Record 0"
	writeErr0 := disk.Write(0, expected0)
	if writeErr0 != nil {
		t.Fatalf(writeErr0.Error())
	}
	expected1 := "Record 1"
	writeErr1 := disk.Write(1, expected1)
	if writeErr1 != nil {
		t.Fatalf(writeErr1.Error())
	}
	actual0, readErr0 := disk.Read(0)
	if readErr0 != nil {
		t.Fatalf(readErr0.Error())
	}
	if actual0 != expected0 {
		t.Fatalf(fmt.Sprintf("Expected: \"%s\", Actual: %s", expected0, actual0))
	}
	actual1, readErr1 := disk.Read(1)
	if readErr1 != nil {
		t.Fatalf(readErr1.Error())
	}
	if actual1 != expected1 {
		t.Fatalf(fmt.Sprintf("Expected: \"%s\", Actual: %s", expected1, actual1))
	}
}

func TestStress(t *testing.T) {
	disk, err := storage.NewStorage("disk3")
	if err != nil {
		t.Fatalf(err.Error())
	}
	for i := int64(0); i < 10240; i++ {
		expected := fmt.Sprintf("Record %d", i)
		writeErr := disk.Write(i, expected)
		if writeErr != nil {
			t.Fatalf(writeErr.Error())
		}
		actual, readErr := disk.Read(i)
		if readErr != nil {
			t.Fatalf(readErr.Error())
		}
		if actual != expected {
			t.Fatalf(fmt.Sprintf("Expected: \"%s\", Actual: %s", expected, actual))
		}
	}
}
