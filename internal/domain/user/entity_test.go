package user

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name      string
		userName  string
		email     string
		wantErr   bool
		errType   error
	}{
		{
			name:     "valid user",
			userName: "John Doe",
			email:    "john@example.com",
			wantErr:  false,
		},
		{
			name:     "invalid email",
			userName: "John Doe",
			email:    "invalid-email",
			wantErr:  true,
			errType:  ErrInvalidEmail,
		},
		{
			name:     "empty email",
			userName: "John Doe",
			email:    "",
			wantErr:  true,
			errType:  ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.userName, tt.email)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUser() expected error but got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewUser() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("NewUser() unexpected error = %v", err)
				return
			}

			if user == nil {
				t.Error("NewUser() returned nil user")
				return
			}

			if user.Name != tt.userName {
				t.Errorf("NewUser() name = %v, want %v", user.Name, tt.userName)
			}

			if user.Email != tt.email {
				t.Errorf("NewUser() email = %v, want %v", user.Email, tt.email)
			}

			if user.ID == "" {
				t.Error("NewUser() ID should not be empty")
			}

			if user.IsSystem {
				t.Error("NewUser() IsSystem should be false by default")
			}

			if user.IsAdmin {
				t.Error("NewUser() IsAdmin should be false by default")
			}

			if user.CreatedTime.IsZero() {
				t.Error("NewUser() CreatedTime should not be zero")
			}
		})
	}
}

func TestNewUserWithPassword(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		email    string
		password string
		wantErr  bool
		errType  error
	}{
		{
			name:     "valid user with strong password",
			userName: "John Doe",
			email:    "john@example.com",
			password: "strongPassword123",
			wantErr:  false,
		},
		{
			name:     "weak password",
			userName: "John Doe",
			email:    "john@example.com",
			password: "weak",
			wantErr:  true,
			errType:  ErrWeakPassword,
		},
		{
			name:     "password without numbers",
			userName: "John Doe",
			email:    "john@example.com",
			password: "onlyletters",
			wantErr:  true,
			errType:  ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUserWithPassword(tt.userName, tt.email, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUserWithPassword() expected error but got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewUserWithPassword() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("NewUserWithPassword() unexpected error = %v", err)
				return
			}

			if user.Password == nil {
				t.Error("NewUserWithPassword() password should not be nil")
			}

			// 验证密码
			if err := user.CheckPassword(tt.password); err != nil {
				t.Errorf("NewUserWithPassword() password verification failed: %v", err)
			}
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	user, _ := NewUser("John Doe", "john@example.com")

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errType  error
	}{
		{
			name:     "valid password",
			password: "validPassword123",
			wantErr:  false,
		},
		{
			name:     "weak password",
			password: "weak",
			wantErr:  true,
			errType:  ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.SetPassword(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SetPassword() expected error but got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("SetPassword() error = %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Errorf("SetPassword() unexpected error = %v", err)
				return
			}

			if user.Password == nil {
				t.Error("SetPassword() password should not be nil")
			}

			// 验证密码
			if err := user.CheckPassword(tt.password); err != nil {
				t.Errorf("SetPassword() password verification failed: %v", err)
			}
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user, _ := NewUserWithPassword("John Doe", "john@example.com", "validPassword123")

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: "validPassword123",
			wantErr:  false,
		},
		{
			name:     "wrong password",
			password: "wrongPassword",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.CheckPassword(tt.password)

			if tt.wantErr && err == nil {
				t.Errorf("CheckPassword() expected error but got nil")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("CheckPassword() unexpected error = %v", err)
			}
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	user, _ := NewUser("John Doe", "john@example.com")

	// 初始状态应该是激活的
	if !user.IsActive() {
		t.Error("User should be active by default")
	}

	// 停用用户
	user.Deactivate()
	if user.IsActive() {
		t.Error("User should not be active after deactivation")
	}

	// 重新激活
	user.Activate()
	if !user.IsActive() {
		t.Error("User should be active after activation")
	}

	// 软删除
	user.SoftDelete()
	if user.IsActive() {
		t.Error("User should not be active after soft delete")
	}
}

func TestUser_GetStatus(t *testing.T) {
	user, _ := NewUser("John Doe", "john@example.com")

	// 初始状态
	if user.GetStatus() != UserStatusActive {
		t.Errorf("User status = %v, want %v", user.GetStatus(), UserStatusActive)
	}

	// 停用
	user.Deactivate()
	if user.GetStatus() != UserStatusDeactivated {
		t.Errorf("User status = %v, want %v", user.GetStatus(), UserStatusDeactivated)
	}

	// 软删除
	user.SoftDelete()
	if user.GetStatus() != UserStatusDeleted {
		t.Errorf("User status = %v, want %v", user.GetStatus(), UserStatusDeleted)
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	user, _ := NewUser("John Doe", "john@example.com")
	oldModifiedTime := user.LastModifiedTime

	// 等待一毫秒确保时间不同
	time.Sleep(time.Millisecond)

	newName := "Jane Doe"
	newPhone := "1234567890"
	newAvatar := "https://example.com/avatar.jpg"

	user.UpdateProfile(&newName, &newPhone, &newAvatar)

	if user.Name != newName {
		t.Errorf("UpdateProfile() name = %v, want %v", user.Name, newName)
	}

	if user.Phone == nil || *user.Phone != newPhone {
		t.Errorf("UpdateProfile() phone = %v, want %v", user.Phone, newPhone)
	}

	if user.Avatar == nil || *user.Avatar != newAvatar {
		t.Errorf("UpdateProfile() avatar = %v, want %v", user.Avatar, newAvatar)
	}

	if user.LastModifiedTime == oldModifiedTime {
		t.Error("UpdateProfile() should update LastModifiedTime")
	}
}

func TestUser_PromoteToAdmin(t *testing.T) {
	user, _ := NewUser("John Doe", "john@example.com")

	if user.IsAdmin {
		t.Error("User should not be admin by default")
	}

	user.PromoteToAdmin()

	if !user.IsAdmin {
		t.Error("User should be admin after promotion")
	}
}

func TestUser_DemoteFromAdmin(t *testing.T) {
	user, _ := NewUser("John Doe", "john@example.com")
	user.PromoteToAdmin()

	if !user.IsAdmin {
		t.Error("User should be admin after promotion")
	}

	user.DemoteFromAdmin()

	if user.IsAdmin {
		t.Error("User should not be admin after demotion")
	}
}

func TestUser_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		email    string
		want     string
	}{
		{
			name:     "user with name",
			userName: "John Doe",
			email:    "john@example.com",
			want:     "John Doe",
		},
		{
			name:     "user without name",
			userName: "",
			email:    "john@example.com",
			want:     "john@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, _ := NewUser(tt.userName, tt.email)
			if got := user.GetDisplayName(); got != tt.want {
				t.Errorf("GetDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_AddAccount(t *testing.T) {
	user, _ := NewUser("John Doe", "john@example.com")

	account := user.AddAccount(AccountTypeOAuth, ProviderGitHub, "github123")

	if account == nil {
		t.Error("AddAccount() should return account")
	}

	if account.UserID != user.ID {
		t.Errorf("AddAccount() account.UserID = %v, want %v", account.UserID, user.ID)
	}

	if account.Type != string(AccountTypeOAuth) {
		t.Errorf("AddAccount() account.Type = %v, want %v", account.Type, AccountTypeOAuth)
	}

	if account.Provider != string(ProviderGitHub) {
		t.Errorf("AddAccount() account.Provider = %v, want %v", account.Provider, ProviderGitHub)
	}

	if account.ProviderID != "github123" {
		t.Errorf("AddAccount() account.ProviderID = %v, want %v", account.ProviderID, "github123")
	}
}

// 测试辅助函数
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		want  bool
	}{
		{"test@example.com", true},
		{"user@domain.org", true},
		{"invalid-email", false},
		{"", false},
		{"@example.com", false},
		{"test@", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := isValidEmail(tt.email); got != tt.want {
				t.Errorf("isValidEmail(%v) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		password string
		want     bool
	}{
		{"validPassword123", true},
		{"Password1", true},
		{"12345678", false}, // 只有数字
		{"password", false}, // 只有字母
		{"short1", false},   // 太短
		{"", false},         // 空密码
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			if got := isValidPassword(tt.password); got != tt.want {
				t.Errorf("isValidPassword(%v) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}