package directory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/directory_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/group"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/member"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/user"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/user/name"
)

func testServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return NewClient(directory_config.WithBaseUrl(u))
}

// User operations

func TestCreateUser(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input user.User
		json.NewDecoder(r.Body).Decode(&input)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&user.User{
			Kind:         "admin#directory#user",
			Id:           "123",
			PrimaryEmail: input.PrimaryEmail,
			Name:         input.Name,
		})
	})

	u, err := client.CreateUser(context.Background(), &user.User{
		PrimaryEmail: "test@example.com",
		Name:         &name.Name{GivenName: "Test", FamilyName: "User"},
		Password:     "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.PrimaryEmail != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", u.PrimaryEmail)
	}
	if u.Id != "123" {
		t.Errorf("expected id '123', got %q", u.Id)
	}
}

func TestCreateUser_NilUser(t *testing.T) {
	client := NewClient()
	u, err := client.CreateUser(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != nil {
		t.Error("expected nil for nil user")
	}
}

func TestCreateUser_CancelledContext(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.CreateUser(ctx, &user.User{PrimaryEmail: "test@example.com"})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestGetUser(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/users/test@example.com") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&user.User{
			Kind:         "admin#directory#user",
			Id:           "123",
			PrimaryEmail: "test@example.com",
		})
	})

	u, err := client.GetUser(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.PrimaryEmail != "test@example.com" {
		t.Errorf("expected 'test@example.com', got %q", u.PrimaryEmail)
	}
}

func TestGetUser_EmptyKey(t *testing.T) {
	client := NewClient()
	_, err := client.GetUser(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty user key")
	}
}

func TestUpdateUser(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&user.User{
			Id:           "123",
			PrimaryEmail: "test@example.com",
			Suspended:    true,
		})
	})

	u, err := client.UpdateUser(context.Background(), "test@example.com", &user.User{Suspended: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !u.Suspended {
		t.Error("expected user to be suspended")
	}
}

func TestUpdateUser_EmptyKey(t *testing.T) {
	client := NewClient()
	_, err := client.UpdateUser(context.Background(), "", &user.User{})
	if err == nil {
		t.Fatal("expected error for empty user key")
	}
}

func TestUpdateUser_NilUser(t *testing.T) {
	client := NewClient()
	u, err := client.UpdateUser(context.Background(), "key", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != nil {
		t.Error("expected nil for nil user")
	}
}

func TestDeleteUser(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/users/test@example.com") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteUser(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteUser_EmptyKey(t *testing.T) {
	client := NewClient()
	err := client.DeleteUser(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty user key")
	}
}

// Group operations

func TestCreateGroup(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/groups") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input group.Group
		json.NewDecoder(r.Body).Decode(&input)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&group.Group{
			Kind:  "admin#directory#group",
			Id:    "g-123",
			Email: input.Email,
			Name:  input.Name,
		})
	})

	g, err := client.CreateGroup(context.Background(), &group.Group{
		Email: "group@example.com",
		Name:  "Test Group",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Email != "group@example.com" {
		t.Errorf("expected email 'group@example.com', got %q", g.Email)
	}
}

func TestCreateGroup_NilGroup(t *testing.T) {
	client := NewClient()
	g, err := client.CreateGroup(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g != nil {
		t.Error("expected nil for nil group")
	}
}

func TestGetGroup(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&group.Group{
			Kind:  "admin#directory#group",
			Email: "group@example.com",
		})
	})

	g, err := client.GetGroup(context.Background(), "group@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Email != "group@example.com" {
		t.Errorf("expected 'group@example.com', got %q", g.Email)
	}
}

func TestGetGroup_EmptyKey(t *testing.T) {
	client := NewClient()
	_, err := client.GetGroup(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty group key")
	}
}

func TestUpdateGroup(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&group.Group{
			Email:       "group@example.com",
			Description: "Updated",
		})
	})

	g, err := client.UpdateGroup(context.Background(), "group@example.com", &group.Group{Description: "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Description != "Updated" {
		t.Errorf("expected description 'Updated', got %q", g.Description)
	}
}

func TestUpdateGroup_EmptyKey(t *testing.T) {
	client := NewClient()
	_, err := client.UpdateGroup(context.Background(), "", &group.Group{})
	if err == nil {
		t.Fatal("expected error for empty group key")
	}
}

func TestUpdateGroup_NilGroup(t *testing.T) {
	client := NewClient()
	g, err := client.UpdateGroup(context.Background(), "key", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g != nil {
		t.Error("expected nil for nil group")
	}
}

func TestDeleteGroup(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteGroup(context.Background(), "group@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteGroup_EmptyKey(t *testing.T) {
	client := NewClient()
	err := client.DeleteGroup(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty group key")
	}
}

// Member operations

func TestCreateMember(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/members") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input member.Member
		json.NewDecoder(r.Body).Decode(&input)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&member.Member{
			Kind:  "admin#directory#member",
			Id:    "m-123",
			Email: input.Email,
			Role:  input.Role,
		})
	})

	m, err := client.CreateMember(context.Background(), "group@example.com", &member.Member{
		Email: "user@example.com",
		Role:  "MEMBER",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got %q", m.Email)
	}
	if m.Role != "MEMBER" {
		t.Errorf("expected role 'MEMBER', got %q", m.Role)
	}
}

func TestCreateMember_EmptyGroupKey(t *testing.T) {
	client := NewClient()
	_, err := client.CreateMember(context.Background(), "", &member.Member{})
	if err == nil {
		t.Fatal("expected error for empty group key")
	}
}

func TestCreateMember_NilMember(t *testing.T) {
	client := NewClient()
	m, err := client.CreateMember(context.Background(), "group@example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != nil {
		t.Error("expected nil for nil member")
	}
}

func TestGetMember(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&member.Member{
			Email: "user@example.com",
			Role:  "OWNER",
		})
	})

	m, err := client.GetMember(context.Background(), "group@example.com", "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Role != "OWNER" {
		t.Errorf("expected role 'OWNER', got %q", m.Role)
	}
}

func TestGetMember_EmptyGroupKey(t *testing.T) {
	client := NewClient()
	_, err := client.GetMember(context.Background(), "", "member")
	if err == nil {
		t.Fatal("expected error for empty group key")
	}
}

func TestGetMember_EmptyMemberKey(t *testing.T) {
	client := NewClient()
	_, err := client.GetMember(context.Background(), "group", "")
	if err == nil {
		t.Fatal("expected error for empty member key")
	}
}

func TestUpdateMember(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&member.Member{
			Email: "user@example.com",
			Role:  "MANAGER",
		})
	})

	m, err := client.UpdateMember(context.Background(), "group@example.com", "user@example.com", &member.Member{Role: "MANAGER"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Role != "MANAGER" {
		t.Errorf("expected role 'MANAGER', got %q", m.Role)
	}
}

func TestUpdateMember_EmptyGroupKey(t *testing.T) {
	client := NewClient()
	_, err := client.UpdateMember(context.Background(), "", "member", &member.Member{})
	if err == nil {
		t.Fatal("expected error for empty group key")
	}
}

func TestUpdateMember_EmptyMemberKey(t *testing.T) {
	client := NewClient()
	_, err := client.UpdateMember(context.Background(), "group", "", &member.Member{})
	if err == nil {
		t.Fatal("expected error for empty member key")
	}
}

func TestUpdateMember_NilMember(t *testing.T) {
	client := NewClient()
	m, err := client.UpdateMember(context.Background(), "group", "member", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m != nil {
		t.Error("expected nil for nil member")
	}
}

func TestDeleteMember(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteMember(context.Background(), "group@example.com", "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMember_EmptyGroupKey(t *testing.T) {
	client := NewClient()
	err := client.DeleteMember(context.Background(), "", "member")
	if err == nil {
		t.Fatal("expected error for empty group key")
	}
}

func TestDeleteMember_EmptyMemberKey(t *testing.T) {
	client := NewClient()
	err := client.DeleteMember(context.Background(), "group", "")
	if err == nil {
		t.Fatal("expected error for empty member key")
	}
}
