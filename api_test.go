package main

import (
	"fmt"
	"testing"
)

func TestMemDB(t *testing.T) {
	// Create a new instance of your key-value store
	db := NewInMem()

	// Test Set functionality
	testKey := []byte("testKey")
	testValue := []byte("testValue")
	db.Set(testKey, testValue)

	// Test Get functionality
	result, err := db.Get(testKey)
	if err != nil {
		t.Fatalf("Error getting value for key %s: %v", testKey, err)
	}

	resultString := string(result)
	testValueString := string(testValue)
	if resultString != testValueString {
		t.Errorf("Expected value %s, got %s", testValue, result)
	}

	// Test Delete functionality
	deletedValue, err := db.Del(string(testKey))
	if err != nil {
		t.Fatalf("Error deleting key %s: %v", testKey, err)
	}
	if deletedValue != testValueString {
		t.Errorf("Expected deleted value %s, got %s", testValue, deletedValue)
	}

	// Verify that the key is no longer present in the database
	_, err = db.Get(testKey)
	if err == nil {
		t.Errorf("Expected key %s to be deleted, but it still exists", testKey)
	}

	fmt.Println("TestMemDB : ok")
}

func TestMemDB_SetGet(t *testing.T) {
	db := NewInMem()

	testKey := []byte("testKey")
	testValue := []byte("testValue")

	// Test Set functionality
	err := db.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Error setting value for key %s: %v", testKey, err)
	}

	// Test Get functionality
	result, err := db.Get(testKey)
	if err != nil {
		t.Fatalf("Error getting value for key %s: %v", testKey, err)
	}

	resultString := string(result)
	testValueString := string(testValue)
	if resultString != testValueString {
		t.Errorf("Expected value %s, got %s", testValueString, resultString)
	}

	fmt.Println("TestMemDB_SetGet : ok")
}

func TestMemDB_SetDel(t *testing.T) {
	db := NewInMem()

	testKey := []byte("testKey")
	testValue := []byte("testValue")

	// Test Set functionality
	err := db.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Error setting value for key %s: %v", testKey, err)
	}

	// Test Delete functionality
	_, err = db.Del(string(testKey))
	if err != nil {
		t.Fatalf("Error deleting key %s: %v", testKey, err)
	}

	// Verify that the key is no longer present in the database
	_, err = db.Get(testKey)
	if err == nil {
		t.Errorf("Expected key %s to be deleted, but it still exists", testKey)
	}

	fmt.Println("TestMemDB_SetDel : ok")
}

func TestMemDB_SetGetDel(t *testing.T) {
	db := NewInMem()

	testKey := []byte("testKey")
	testValue := []byte("testValue")

	// Test Set functionality
	err := db.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Error setting value for key %s: %v", testKey, err)
	}

	// Test Get functionality
	result, err := db.Get(testKey)
	if err != nil {
		t.Fatalf("Error getting value for key %s: %v", testKey, err)
	}

	resultString := string(result)
	testValueString := string(testValue)
	if resultString != testValueString {
		t.Errorf("Expected value %s, got %s", testValueString, resultString)
	}

	// Test Delete functionality
	_, err = db.Del(string(testKey))
	if err != nil {
		t.Fatalf("Error deleting key %s: %v", testKey, err)
	}

	// Verify that the key is no longer present in the database
	_, err = db.Get(testKey)
	if err == nil {
		t.Errorf("Expected key %s to be deleted, but it still exists", testKey)
	}
	fmt.Println("TestMemDB_SetGetDel : ok")
}
