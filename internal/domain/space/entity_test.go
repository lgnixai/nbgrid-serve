package space

import (
	"testing"
	"time"
)

func TestNewSpace(t *testing.T) {
	// Test valid space creation
	space, err := NewSpace("Test Space", "user123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if space.Name != "Test Space" {
		t.Errorf("Expected name 'Test Space', got %s", space.Name)
	}
	
	if space.CreatedBy != "user123" {
		t.Errorf("Expected created by 'user123', got %s", space.CreatedBy)
	}
	
	if space.GetStatus() != SpaceStatusActive {
		t.Errorf("Expected status active, got %s", space.GetStatus())
	}
	
	if space.GetMemberCount() != 1 {
		t.Errorf("Expected member count 1, got %d", space.GetMemberCount())
	}
}

func TestSpaceValidation(t *testing.T) {
	// Test empty name
	_, err := NewSpace("", "user123")
	if err == nil {
		t.Error("Expected error for empty name")
	}
	
	// Test empty created by
	_, err = NewSpace("Test Space", "")
	if err == nil {
		t.Error("Expected error for empty created by")
	}
}

func TestSpaceArchive(t *testing.T) {
	space, _ := NewSpace("Test Space", "user123")
	
	// Test archive
	err := space.Archive()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !space.IsArchived() {
		t.Error("Expected space to be archived")
	}
	
	// Test double archive
	err = space.Archive()
	if err == nil {
		t.Error("Expected error for double archive")
	}
}

func TestSpaceRestore(t *testing.T) {
	space, _ := NewSpace("Test Space", "user123")
	space.Archive()
	
	// Test restore
	err := space.Restore()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !space.IsActive() {
		t.Error("Expected space to be active")
	}
}

func TestNewSpaceCollaborator(t *testing.T) {
	// Test valid collaborator creation
	collab, err := NewSpaceCollaborator("space123", "user456", RoleEditor, "user123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if collab.SpaceID != "space123" {
		t.Errorf("Expected space ID 'space123', got %s", collab.SpaceID)
	}
	
	if collab.UserID != "user456" {
		t.Errorf("Expected user ID 'user456', got %s", collab.UserID)
	}
	
	if collab.Role != RoleEditor {
		t.Errorf("Expected role editor, got %s", collab.Role)
	}
	
	if collab.Status != CollaboratorStatusPending {
		t.Errorf("Expected status pending, got %s", collab.Status)
	}
}

func TestCollaboratorAccept(t *testing.T) {
	collab, _ := NewSpaceCollaborator("space123", "user456", RoleEditor, "user123")
	
	// Test accept
	err := collab.Accept()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !collab.IsActive() {
		t.Error("Expected collaborator to be active")
	}
	
	if collab.AcceptedAt == nil {
		t.Error("Expected accepted at to be set")
	}
}

func TestCollaboratorPermissions(t *testing.T) {
	collab, _ := NewSpaceCollaborator("space123", "user456", RoleEditor, "user123")
	collab.Accept()
	
	// Test editor permissions
	if !collab.HasPermission(PermissionRead) {
		t.Error("Expected editor to have read permission")
	}
	
	if !collab.HasPermission(PermissionWrite) {
		t.Error("Expected editor to have write permission")
	}
	
	if collab.HasPermission(PermissionManageMembers) {
		t.Error("Expected editor to not have manage members permission")
	}
}

func TestRolePermissions(t *testing.T) {
	// Test owner permissions
	ownerPerms := RoleOwner.GetPermissions()
	if len(ownerPerms) == 0 {
		t.Error("Expected owner to have permissions")
	}
	
	// Test viewer permissions
	viewerPerms := RoleViewer.GetPermissions()
	if len(viewerPerms) != 1 || viewerPerms[0] != PermissionRead {
		t.Error("Expected viewer to only have read permission")
	}
}

func TestRoleManagement(t *testing.T) {
	// Test owner can manage admin
	if !RoleOwner.CanManageRole(RoleAdmin) {
		t.Error("Expected owner to manage admin")
	}
	
	// Test admin cannot manage owner
	if RoleAdmin.CanManageRole(RoleOwner) {
		t.Error("Expected admin to not manage owner")
	}
	
	// Test editor cannot manage viewer
	if RoleEditor.CanManageRole(RoleViewer) {
		t.Error("Expected editor to not manage viewer")
	}
}