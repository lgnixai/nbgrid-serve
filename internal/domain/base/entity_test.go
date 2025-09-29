package base

import (
	"testing"
)

func TestNewBase(t *testing.T) {
	// Test valid base creation
	base, err := NewBase("space123", "Test Base", "user123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if base.Name != "Test Base" {
		t.Errorf("Expected name 'Test Base', got %s", base.Name)
	}
	
	if base.SpaceID != "space123" {
		t.Errorf("Expected space ID 'space123', got %s", base.SpaceID)
	}
	
	if base.CreatedBy != "user123" {
		t.Errorf("Expected created by 'user123', got %s", base.CreatedBy)
	}
	
	if base.GetStatus() != BaseStatusActive {
		t.Errorf("Expected status active, got %s", base.GetStatus())
	}
	
	if base.GetTableCount() != 0 {
		t.Errorf("Expected table count 0, got %d", base.GetTableCount())
	}
}

func TestBaseValidation(t *testing.T) {
	// Test empty space ID
	_, err := NewBase("", "Test Base", "user123")
	if err == nil {
		t.Error("Expected error for empty space ID")
	}
	
	// Test empty name
	_, err = NewBase("space123", "", "user123")
	if err == nil {
		t.Error("Expected error for empty name")
	}
	
	// Test empty created by
	_, err = NewBase("space123", "Test Base", "")
	if err == nil {
		t.Error("Expected error for empty created by")
	}
}

func TestBaseArchive(t *testing.T) {
	base, _ := NewBase("space123", "Test Base", "user123")
	
	// Test archive
	err := base.Archive()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !base.IsArchived() {
		t.Error("Expected base to be archived")
	}
	
	// Test double archive
	err = base.Archive()
	if err == nil {
		t.Error("Expected error for double archive")
	}
}

func TestBaseRestore(t *testing.T) {
	base, _ := NewBase("space123", "Test Base", "user123")
	base.Archive()
	
	// Test restore
	err := base.Restore()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !base.IsActive() {
		t.Error("Expected base to be active")
	}
}

func TestBaseUpdate(t *testing.T) {
	base, _ := NewBase("space123", "Test Base", "user123")
	
	newName := "Updated Base"
	newDesc := "Updated description"
	
	err := base.Update(&newName, &newDesc, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if base.Name != newName {
		t.Errorf("Expected name '%s', got %s", newName, base.Name)
	}
	
	if base.Description == nil || *base.Description != newDesc {
		t.Errorf("Expected description '%s', got %v", newDesc, base.Description)
	}
}

func TestBaseValidateForUpdate(t *testing.T) {
	base, _ := NewBase("space123", "Test Base", "user123")
	
	// Test valid update
	err := base.ValidateForUpdate()
	if err != nil {
		t.Errorf("Expected no error for valid update, got %v", err)
	}
	
	// Test update on deleted base
	base.SoftDelete()
	err = base.ValidateForUpdate()
	if err == nil {
		t.Error("Expected error for updating deleted base")
	}
}

func TestBaseValidateForDeletion(t *testing.T) {
	base, _ := NewBase("space123", "Test Base", "user123")
	
	// Test valid deletion
	err := base.ValidateForDeletion()
	if err != nil {
		t.Errorf("Expected no error for valid deletion, got %v", err)
	}
	
	// Test deletion of system base
	base.IsSystem = true
	err = base.ValidateForDeletion()
	if err == nil {
		t.Error("Expected error for deleting system base")
	}
}

func TestBaseTableCount(t *testing.T) {
	base, _ := NewBase("space123", "Test Base", "user123")
	
	// Test initial count
	if base.GetTableCount() != 0 {
		t.Errorf("Expected initial table count 0, got %d", base.GetTableCount())
	}
	
	// Test update count
	base.UpdateTableCount(5)
	if base.GetTableCount() != 5 {
		t.Errorf("Expected table count 5, got %d", base.GetTableCount())
	}
	
	// Test negative count (should not update)
	base.UpdateTableCount(-1)
	if base.GetTableCount() != 5 {
		t.Errorf("Expected table count to remain 5, got %d", base.GetTableCount())
	}
}