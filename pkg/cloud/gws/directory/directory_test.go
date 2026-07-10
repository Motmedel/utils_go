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
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/list_role_assignments_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/asp"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/group"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/member"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/org_unit"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/privilege"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/role"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/role/role_privilege"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/role_assignment"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/token"
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

// User security operations

func TestMakeUserAdmin(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/test@example.com/makeAdmin") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input struct {
			Status bool `json:"status"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		if !input.Status {
			t.Error("expected status true")
		}

		w.WriteHeader(http.StatusNoContent)
	})

	err := client.MakeUserAdmin(context.Background(), "test@example.com", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMakeUserAdmin_EmptyKey(t *testing.T) {
	client := NewClient()
	err := client.MakeUserAdmin(context.Background(), "", true)
	if err == nil {
		t.Fatal("expected error for empty user key")
	}
}

func TestSignOutUser(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/test@example.com/signOut") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.SignOutUser(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignOutUser_EmptyKey(t *testing.T) {
	client := NewClient()
	err := client.SignOutUser(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty user key")
	}
}

func TestTurnOffUser2Sv(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/test@example.com/twoStepVerification/turnOff") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.TurnOffUser2Sv(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTurnOffUser2Sv_EmptyKey(t *testing.T) {
	client := NewClient()
	err := client.TurnOffUser2Sv(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty user key")
	}
}

// Org unit operations

func TestCreateOrgUnit(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/orgunits") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input org_unit.OrgUnit
		json.NewDecoder(r.Body).Decode(&input)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&org_unit.OrgUnit{
			Kind:              "admin#directory#orgUnit",
			OrgUnitId:         "id:123",
			Name:              input.Name,
			OrgUnitPath:       "/" + input.Name,
			ParentOrgUnitPath: input.ParentOrgUnitPath,
		})
	})

	ou, err := client.CreateOrgUnit(context.Background(), "my_customer", &org_unit.OrgUnit{
		Name:              "Engineering",
		ParentOrgUnitPath: "/",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ou.OrgUnitPath != "/Engineering" {
		t.Errorf("expected org unit path '/Engineering', got %q", ou.OrgUnitPath)
	}
	if ou.OrgUnitId != "id:123" {
		t.Errorf("expected org unit id 'id:123', got %q", ou.OrgUnitId)
	}
}

func TestCreateOrgUnit_NilOrgUnit(t *testing.T) {
	client := NewClient()
	ou, err := client.CreateOrgUnit(context.Background(), "my_customer", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ou != nil {
		t.Error("expected nil for nil org unit")
	}
}

func TestCreateOrgUnit_EmptyCustomer(t *testing.T) {
	client := NewClient()
	_, err := client.CreateOrgUnit(context.Background(), "", &org_unit.OrgUnit{Name: "Engineering"})
	if err == nil {
		t.Fatal("expected error for empty customer")
	}
}

func TestGetOrgUnit(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/orgunits/Engineering/Frontend") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&org_unit.OrgUnit{
			Kind:        "admin#directory#orgUnit",
			Name:        "Frontend",
			OrgUnitPath: "/Engineering/Frontend",
		})
	})

	ou, err := client.GetOrgUnit(context.Background(), "my_customer", "/Engineering/Frontend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ou.OrgUnitPath != "/Engineering/Frontend" {
		t.Errorf("expected org unit path '/Engineering/Frontend', got %q", ou.OrgUnitPath)
	}
}

func TestGetOrgUnit_EmptyCustomer(t *testing.T) {
	client := NewClient()
	_, err := client.GetOrgUnit(context.Background(), "", "/Engineering")
	if err == nil {
		t.Fatal("expected error for empty customer")
	}
}

func TestGetOrgUnit_EmptyPath(t *testing.T) {
	client := NewClient()
	_, err := client.GetOrgUnit(context.Background(), "my_customer", "")
	if err == nil {
		t.Fatal("expected error for empty org unit path")
	}
}

func TestUpdateOrgUnit(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/orgunits/Engineering") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input org_unit.OrgUnit
		json.NewDecoder(r.Body).Decode(&input)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&input)
	})

	ou, err := client.UpdateOrgUnit(context.Background(), "my_customer", "/Engineering", &org_unit.OrgUnit{
		Name:        "Engineering",
		Description: "Updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ou.Description != "Updated" {
		t.Errorf("expected description 'Updated', got %q", ou.Description)
	}
}

func TestUpdateOrgUnit_NilOrgUnit(t *testing.T) {
	client := NewClient()
	ou, err := client.UpdateOrgUnit(context.Background(), "my_customer", "/Engineering", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ou != nil {
		t.Error("expected nil for nil org unit")
	}
}

func TestDeleteOrgUnit(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/orgunits/Engineering") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteOrgUnit(context.Background(), "my_customer", "/Engineering")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListOrgUnits(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("type"); got != "all" {
			t.Errorf("expected type 'all', got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"organizationUnits": []*org_unit.OrgUnit{
				{Name: "Engineering", OrgUnitPath: "/Engineering"},
				{Name: "Frontend", OrgUnitPath: "/Engineering/Frontend"},
			},
		})
	})

	orgUnits, err := client.ListOrgUnits(context.Background(), "my_customer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orgUnits) != 2 {
		t.Fatalf("expected 2 org units, got %d", len(orgUnits))
	}
	if orgUnits[1].OrgUnitPath != "/Engineering/Frontend" {
		t.Errorf("expected org unit path '/Engineering/Frontend', got %q", orgUnits[1].OrgUnitPath)
	}
}

// Role operations

func TestListRoles_Pagination(t *testing.T) {
	requestCount := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("pageToken") == "" {
			json.NewEncoder(w).Encode(map[string]any{
				"items":         []*role.Role{{RoleId: "1", RoleName: "_SEED_ADMIN_ROLE"}},
				"nextPageToken": "next",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]any{
				"items": []*role.Role{{RoleId: "2", RoleName: "_GROUPS_ADMIN_ROLE"}},
			})
		}
	})

	roles, err := client.ListRoles(context.Background(), "my_customer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 2 {
		t.Errorf("expected 2 requests, got %d", requestCount)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if roles[1].RoleId != "2" {
		t.Errorf("expected role id '2', got %q", roles[1].RoleId)
	}
}

func TestCreateRole(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/roles") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input role.Role
		json.NewDecoder(r.Body).Decode(&input)
		if len(input.RolePrivileges) != 1 || input.RolePrivileges[0].PrivilegeName != "USERS_RETRIEVE" {
			t.Errorf("unexpected role privileges: %+v", input.RolePrivileges)
		}

		input.RoleId = "123"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&input)
	})

	r, err := client.CreateRole(context.Background(), "my_customer", &role.Role{
		RoleName: "User Reader",
		RolePrivileges: []*role_privilege.RolePrivilege{
			{PrivilegeName: "USERS_RETRIEVE", ServiceId: "00haapch16h1ysv"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.RoleId != "123" {
		t.Errorf("expected role id '123', got %q", r.RoleId)
	}
}

func TestCreateRole_NilRole(t *testing.T) {
	client := NewClient()
	r, err := client.CreateRole(context.Background(), "my_customer", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r != nil {
		t.Error("expected nil for nil role")
	}
}

func TestGetRole(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/roles/123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&role.Role{RoleId: "123", RoleName: "User Reader"})
	})

	r, err := client.GetRole(context.Background(), "my_customer", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.RoleName != "User Reader" {
		t.Errorf("expected role name 'User Reader', got %q", r.RoleName)
	}
}

func TestGetRole_EmptyRoleId(t *testing.T) {
	client := NewClient()
	_, err := client.GetRole(context.Background(), "my_customer", "")
	if err == nil {
		t.Fatal("expected error for empty role id")
	}
}

func TestUpdateRole(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/roles/123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input role.Role
		json.NewDecoder(r.Body).Decode(&input)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&input)
	})

	r, err := client.UpdateRole(context.Background(), "my_customer", "123", &role.Role{
		RoleId:   "123",
		RoleName: "User Reader Updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.RoleName != "User Reader Updated" {
		t.Errorf("expected role name 'User Reader Updated', got %q", r.RoleName)
	}
}

func TestDeleteRole(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/roles/123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteRole(context.Background(), "my_customer", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListPrivileges(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/roles/ALL/privileges") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []*privilege.Privilege{
				{
					PrivilegeName: "USERS_ALL",
					ServiceId:     "00haapch16h1ysv",
					IsOuScopable:  true,
					ChildPrivileges: []*privilege.Privilege{
						{PrivilegeName: "USERS_RETRIEVE", ServiceId: "00haapch16h1ysv"},
					},
				},
			},
		})
	})

	privileges, err := client.ListPrivileges(context.Background(), "my_customer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(privileges) != 1 {
		t.Fatalf("expected 1 privilege, got %d", len(privileges))
	}
	if len(privileges[0].ChildPrivileges) != 1 || privileges[0].ChildPrivileges[0].PrivilegeName != "USERS_RETRIEVE" {
		t.Errorf("unexpected child privileges: %+v", privileges[0].ChildPrivileges)
	}
}

// Role assignment operations

func TestListRoleAssignments(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/roleassignments") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("userKey"); got != "test@example.com" {
			t.Errorf("expected user key 'test@example.com', got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []*role_assignment.RoleAssignment{
				{RoleAssignmentId: "1", RoleId: "123", AssignedTo: "user-id", ScopeType: "CUSTOMER"},
			},
		})
	})

	roleAssignments, err := client.ListRoleAssignments(
		context.Background(),
		"my_customer",
		list_role_assignments_config.WithUserKey("test@example.com"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roleAssignments) != 1 {
		t.Fatalf("expected 1 role assignment, got %d", len(roleAssignments))
	}
	if roleAssignments[0].RoleId != "123" {
		t.Errorf("expected role id '123', got %q", roleAssignments[0].RoleId)
	}
}

func TestCreateRoleAssignment(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/roleassignments") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input role_assignment.RoleAssignment
		json.NewDecoder(r.Body).Decode(&input)
		if input.Condition == "" {
			t.Error("expected condition to be set")
		}

		input.RoleAssignmentId = "1"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&input)
	})

	ra, err := client.CreateRoleAssignment(context.Background(), "my_customer", &role_assignment.RoleAssignment{
		RoleId:       "123",
		AssignedTo:   "user-id",
		AssigneeType: "user",
		ScopeType:    "CUSTOMER",
		Condition:    "!api.getAttribute('cloudidentity.googleapis.com/groups.labels', []).hasAny(['groups.security']) && resource.type == 'cloudidentity.googleapis.com/Group'",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ra.RoleAssignmentId != "1" {
		t.Errorf("expected role assignment id '1', got %q", ra.RoleAssignmentId)
	}
}

func TestCreateRoleAssignment_NilRoleAssignment(t *testing.T) {
	client := NewClient()
	ra, err := client.CreateRoleAssignment(context.Background(), "my_customer", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ra != nil {
		t.Error("expected nil for nil role assignment")
	}
}

func TestGetRoleAssignment(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/roleassignments/1") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&role_assignment.RoleAssignment{
			RoleAssignmentId: "1",
			RoleId:           "123",
			ScopeType:        "ORG_UNIT",
			OrgUnitId:        "id:456",
		})
	})

	ra, err := client.GetRoleAssignment(context.Background(), "my_customer", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ra.ScopeType != "ORG_UNIT" {
		t.Errorf("expected scope type 'ORG_UNIT', got %q", ra.ScopeType)
	}
}

func TestDeleteRoleAssignment(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/customer/my_customer/roleassignments/1") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteRoleAssignment(context.Background(), "my_customer", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRoleAssignment_EmptyId(t *testing.T) {
	client := NewClient()
	err := client.DeleteRoleAssignment(context.Background(), "my_customer", "")
	if err == nil {
		t.Fatal("expected error for empty role assignment id")
	}
}

// Token operations

func TestListTokens(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/test@example.com/tokens") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []*token.Token{
				{ClientId: "abc.apps.googleusercontent.com", DisplayText: "Some App", Scopes: []string{"https://mail.google.com/"}},
			},
		})
	})

	tokens, err := client.ListTokens(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].ClientId != "abc.apps.googleusercontent.com" {
		t.Errorf("expected client id 'abc.apps.googleusercontent.com', got %q", tokens[0].ClientId)
	}
}

func TestListTokens_EmptyKey(t *testing.T) {
	client := NewClient()
	_, err := client.ListTokens(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty user key")
	}
}

func TestGetToken(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/test@example.com/tokens/abc.apps.googleusercontent.com") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&token.Token{
			ClientId:    "abc.apps.googleusercontent.com",
			DisplayText: "Some App",
			NativeApp:   true,
		})
	})

	tok, err := client.GetToken(context.Background(), "test@example.com", "abc.apps.googleusercontent.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tok.NativeApp {
		t.Error("expected native app true")
	}
}

func TestDeleteToken(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/test@example.com/tokens/abc.apps.googleusercontent.com") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteToken(context.Background(), "test@example.com", "abc.apps.googleusercontent.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteToken_EmptyClientId(t *testing.T) {
	client := NewClient()
	err := client.DeleteToken(context.Background(), "test@example.com", "")
	if err == nil {
		t.Fatal("expected error for empty client id")
	}
}

// Application-specific password operations

func TestListAsps(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/test@example.com/asps") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []*asp.Asp{
				{CodeId: 1, Name: "Mail on old phone"},
			},
		})
	})

	asps, err := client.ListAsps(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(asps) != 1 {
		t.Fatalf("expected 1 asp, got %d", len(asps))
	}
	if asps[0].CodeId != 1 {
		t.Errorf("expected code id 1, got %d", asps[0].CodeId)
	}
}

func TestGetAsp(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/test@example.com/asps/1") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&asp.Asp{CodeId: 1, Name: "Mail on old phone"})
	})

	a, err := client.GetAsp(context.Background(), "test@example.com", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name != "Mail on old phone" {
		t.Errorf("expected name 'Mail on old phone', got %q", a.Name)
	}
}

func TestDeleteAsp(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/users/test@example.com/asps/1") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteAsp(context.Background(), "test@example.com", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAsp_EmptyKey(t *testing.T) {
	client := NewClient()
	err := client.DeleteAsp(context.Background(), "", 1)
	if err == nil {
		t.Fatal("expected error for empty user key")
	}
}
