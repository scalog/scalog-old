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
	lsn, writeErr := disk.Write(expected)
	if writeErr != nil {
		t.Fatalf(writeErr.Error())
	}
	segmentSyncErr := disk.Sync()
	if segmentSyncErr != nil {
		t.Fatalf(segmentSyncErr.Error())
	}
	gsn := int64(0)
	commitErr := disk.Commit(lsn, gsn)
	if commitErr != nil {
		t.Fatalf(commitErr.Error())
	}
	globalIndexSyncErr := disk.Sync()
	if globalIndexSyncErr != nil {
		t.Fatalf(globalIndexSyncErr.Error())
	}
	actual, readErr := disk.Read(lsn)
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
	lsn0, writeErr0 := disk.Write(expected0)
	if writeErr0 != nil {
		t.Fatalf(writeErr0.Error())
	}
	expected1 := "Record 1"
	lsn1, writeErr1 := disk.Write(expected1)
	if writeErr1 != nil {
		t.Fatalf(writeErr1.Error())
	}
	segmentSyncErr := disk.Sync()
	if segmentSyncErr != nil {
		t.Fatalf(segmentSyncErr.Error())
	}
	gsn0 := int64(0)
	commitErr0 := disk.Commit(lsn0, gsn0)
	if commitErr0 != nil {
		t.Fatalf(commitErr0.Error())
	}
	gsn1 := int64(1)
	commitErr1 := disk.Commit(lsn1, gsn1)
	if commitErr1 != nil {
		t.Fatalf(commitErr1.Error())
	}
	globalIndexSyncErr := disk.Sync()
	if globalIndexSyncErr != nil {
		t.Fatalf(globalIndexSyncErr.Error())
	}
	actual0, readErr0 := disk.Read(lsn0)
	if readErr0 != nil {
		t.Fatalf(readErr0.Error())
	}
	if actual0 != expected0 {
		t.Fatalf(fmt.Sprintf("Expected: \"%s\", Actual: %s", expected0, actual0))
	}
	actual1, readErr1 := disk.Read(lsn1)
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
	lsnToExpected := make(map[int64]string)
	for i := int64(0); i < 10240; i++ {
		expected := fmt.Sprintf("Record %d", i)
		lsn, writeErr := disk.Write(expected)
		if writeErr != nil {
			t.Fatalf(writeErr.Error())
		}
		lsnToExpected[lsn] = expected
	}
	segmentSyncErr := disk.Sync()
	if segmentSyncErr != nil {
		t.Fatalf(segmentSyncErr.Error())
	}
	for lsn := range lsnToExpected {
		commitErr := disk.Commit(lsn, lsn)
		if commitErr != nil {
			t.Fatalf(commitErr.Error())
		}
	}
	globalIndexSyncErr := disk.Sync()
	if globalIndexSyncErr != nil {
		t.Fatalf(globalIndexSyncErr.Error())
	}
	for lsn, expected := range lsnToExpected {
		actual, readErr := disk.Read(lsn)
		if readErr != nil {
			t.Fatalf(readErr.Error())
		}
		if actual != expected {
			t.Fatalf(fmt.Sprintf("Expected: \"%s\", Actual: %s", expected, actual))
		}
	}
}
