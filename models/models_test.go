package models

import "testing"

func TestCanAssignRole(t *testing.T) {
	admin := &User{Role: RoleAdmin}
	manager := &User{Role: RoleManager}
	customer := &User{Role: RoleCustomer}

	if !admin.CanAssignRole(RoleAdmin) {
		t.Error("admin should be able to assign ADMIN")
	}
	if !admin.CanAssignRole(RoleVendor) {
		t.Error("admin should be able to assign VENDOR")
	}
	if manager.CanAssignRole(RoleAdmin) {
		t.Error("manager must not be able to assign ADMIN")
	}
	if manager.CanAssignRole(RoleSuperAdmin) {
		t.Error("manager must not be able to assign SUPERADMIN")
	}
	if !manager.CanAssignRole(RoleVendor) {
		t.Error("manager should be able to assign non-privileged roles")
	}
	if customer.CanAssignRole(RoleVendor) {
		t.Error("non-management users must not assign any role")
	}
}

func TestIsValidOrderStatus(t *testing.T) {
	if !IsValidOrderStatus(StatusPaid) {
		t.Error("PAID should be a valid status")
	}
	if IsValidOrderStatus("NONSENSE") {
		t.Error("NONSENSE should not be a valid status")
	}
	if IsValidOrderStatus("") {
		t.Error("empty string should not be a valid status")
	}
}

func TestSuperAdminIsAdminAndManagement(t *testing.T) {
	su := &User{Role: RoleSuperAdmin}
	if !su.IsAdmin() {
		t.Error("SUPERADMIN should be treated as admin")
	}
	if !su.IsManagement() {
		t.Error("SUPERADMIN should be treated as management")
	}
}
