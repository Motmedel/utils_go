package directory

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"

	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/directory_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/list_role_assignments_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/asp"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/group"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/member"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/org_unit"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/privilege"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/role"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/role_assignment"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/token"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/user"
)

const Domain = "admin.googleapis.com"

var defaultBaseUrl = &url.URL{
	Scheme: "https",
	Host:   Domain,
}

type Client struct {
	baseUrl *url.URL
	config  *directory_config.Config
}

func NewClient(options ...directory_config.Option) *Client {
	config := directory_config.New(options...)
	baseUrl := config.BaseUrl
	if baseUrl == nil {
		baseUrl = defaultBaseUrl
	}
	u := *baseUrl
	u.Path = "/admin/directory/v1/"
	return &Client{baseUrl: &u, config: config}
}

// User operations

// CreateUser creates a new user account.
func (c *Client) CreateUser(ctx context.Context, u *user.User, options ...fetch_config.Option) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if u == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users"
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, createdUser, err := motmedelHttpUtils.FetchJsonWithBody[*user.User](ctx, urlString, u, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdUser, nil
}

// GetUser retrieves a user account identified by userKey (primary email address, alias email address, or unique user ID).
func (c *Client) GetUser(ctx context.Context, userKey string, options ...fetch_config.Option) (*user.User, error) {
	if userKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey)
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, u, err := motmedelHttpUtils.FetchJson[*user.User](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return u, nil
}

// UpdateUser updates a user account identified by userKey.
func (c *Client) UpdateUser(ctx context.Context, userKey string, u *user.User, options ...fetch_config.Option) (*user.User, error) {
	if userKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if u == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPut))
	_, updatedUser, err := motmedelHttpUtils.FetchJsonWithBody[*user.User](ctx, urlString, u, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updatedUser, nil
}

// DeleteUser deletes a user account identified by userKey.
func (c *Client) DeleteUser(ctx context.Context, userKey string, options ...fetch_config.Option) error {
	if userKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

type listUsersResponse struct {
	Users         []*user.User `json:"users"`
	NextPageToken string       `json:"nextPageToken"`
}

// ListUsers retrieves all users for the given customer ID (use "my_customer" for the authenticated account).
func (c *Client) ListUsers(ctx context.Context, customer string, options ...fetch_config.Option) ([]*user.User, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	var allUsers []*user.User
	pageToken := ""

	for {
		urlObj := *c.baseUrl
		urlObj.Path += "users"

		query := url.Values{}
		query.Set("customer", customer)
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		urlObj.RawQuery = query.Encode()
		urlString := urlObj.String()

		paginatedOptions := append(c.config.FetchOptions, options...)
		_, resp, err := motmedelHttpUtils.FetchJson[*listUsersResponse](ctx, urlString, paginatedOptions...)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
		}

		if resp != nil {
			allUsers = append(allUsers, resp.Users...)

			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}

	return allUsers, nil
}

// Group operations

type listGroupsResponse struct {
	Groups        []*group.Group `json:"groups"`
	NextPageToken string         `json:"nextPageToken"`
}

// ListGroups retrieves all groups for the given customer ID (use "my_customer" for the authenticated account).
func (c *Client) ListGroups(ctx context.Context, customer string, options ...fetch_config.Option) ([]*group.Group, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	var allGroups []*group.Group
	pageToken := ""

	for {
		urlObj := *c.baseUrl
		urlObj.Path += "groups"

		query := url.Values{}
		query.Set("customer", customer)
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		urlObj.RawQuery = query.Encode()
		urlString := urlObj.String()

		paginatedOptions := append(c.config.FetchOptions, options...)
		_, resp, err := motmedelHttpUtils.FetchJson[*listGroupsResponse](ctx, urlString, paginatedOptions...)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
		}

		if resp != nil {
			allGroups = append(allGroups, resp.Groups...)

			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}

	return allGroups, nil
}

// CreateGroup creates a new group.
func (c *Client) CreateGroup(ctx context.Context, g *group.Group, options ...fetch_config.Option) (*group.Group, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if g == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups"
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, createdGroup, err := motmedelHttpUtils.FetchJsonWithBody[*group.Group](ctx, urlString, g, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdGroup, nil
}

// GetGroup retrieves a group identified by groupKey (group email address, group alias, or unique group ID).
func (c *Client) GetGroup(ctx context.Context, groupKey string, options ...fetch_config.Option) (*group.Group, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey)
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, g, err := motmedelHttpUtils.FetchJson[*group.Group](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return g, nil
}

// UpdateGroup updates a group identified by groupKey.
func (c *Client) UpdateGroup(ctx context.Context, groupKey string, g *group.Group, options ...fetch_config.Option) (*group.Group, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if g == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPut))
	_, updatedGroup, err := motmedelHttpUtils.FetchJsonWithBody[*group.Group](ctx, urlString, g, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updatedGroup, nil
}

// DeleteGroup deletes a group identified by groupKey.
func (c *Client) DeleteGroup(ctx context.Context, groupKey string, options ...fetch_config.Option) error {
	if groupKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

// Group member operations

type listMembersResponse struct {
	Members       []*member.Member `json:"members"`
	NextPageToken string           `json:"nextPageToken"`
}

// ListMembers retrieves all members of a group identified by groupKey.
func (c *Client) ListMembers(ctx context.Context, groupKey string, options ...fetch_config.Option) ([]*member.Member, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	var allMembers []*member.Member
	pageToken := ""

	for {
		urlObj := *c.baseUrl
		urlObj.Path += "groups/" + url.PathEscape(groupKey) + "/members"

		query := url.Values{}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		urlObj.RawQuery = query.Encode()
		urlString := urlObj.String()

		paginatedOptions := append(c.config.FetchOptions, options...)
		_, resp, err := motmedelHttpUtils.FetchJson[*listMembersResponse](ctx, urlString, paginatedOptions...)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
		}

		if resp != nil {
			allMembers = append(allMembers, resp.Members...)

			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}

	return allMembers, nil
}

// CreateMember adds a member to a group identified by groupKey.
func (c *Client) CreateMember(ctx context.Context, groupKey string, m *member.Member, options ...fetch_config.Option) (*member.Member, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if m == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey) + "/members"
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, createdMember, err := motmedelHttpUtils.FetchJsonWithBody[*member.Member](ctx, urlString, m, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdMember, nil
}

// GetMember retrieves a member of a group identified by groupKey and memberKey (member email address or unique member ID).
func (c *Client) GetMember(ctx context.Context, groupKey string, memberKey string, options ...fetch_config.Option) (*member.Member, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}
	if memberKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("member key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey) + "/members/" + url.PathEscape(memberKey)
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, m, err := motmedelHttpUtils.FetchJson[*member.Member](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return m, nil
}

// UpdateMember updates a member of a group identified by groupKey and memberKey.
func (c *Client) UpdateMember(ctx context.Context, groupKey string, memberKey string, m *member.Member, options ...fetch_config.Option) (*member.Member, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}
	if memberKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("member key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if m == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey) + "/members/" + url.PathEscape(memberKey)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPut))
	_, updatedMember, err := motmedelHttpUtils.FetchJsonWithBody[*member.Member](ctx, urlString, m, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updatedMember, nil
}

// DeleteMember removes a member from a group identified by groupKey and memberKey.
func (c *Client) DeleteMember(ctx context.Context, groupKey string, memberKey string, options ...fetch_config.Option) error {
	if groupKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}
	if memberKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("member key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey) + "/members/" + url.PathEscape(memberKey)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

// User security operations

type makeAdminRequest struct {
	Status bool `json:"status"`
}

// MakeUserAdmin grants or revokes super administrator status for the user identified by userKey.
func (c *Client) MakeUserAdmin(ctx context.Context, userKey string, status bool, options ...fetch_config.Option) error {
	if userKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey) + "/makeAdmin"
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, _, err := motmedelHttpUtils.FetchJsonWithBody[*struct{}](ctx, urlString, &makeAdminRequest{Status: status}, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return nil
}

// SignOutUser signs the user identified by userKey out of all web and device sessions and resets their sign-in cookies.
func (c *Client) SignOutUser(ctx context.Context, userKey string, options ...fetch_config.Option) error {
	if userKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey) + "/signOut"
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

// TurnOffUser2Sv turns off 2-step verification for the user identified by userKey.
func (c *Client) TurnOffUser2Sv(ctx context.Context, userKey string, options ...fetch_config.Option) error {
	if userKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey) + "/twoStepVerification/turnOff"
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

// Org unit operations

// CreateOrgUnit creates a new organizational unit for the given customer ID (use "my_customer" for the authenticated account).
func (c *Client) CreateOrgUnit(ctx context.Context, customer string, ou *org_unit.OrgUnit, options ...fetch_config.Option) (*org_unit.OrgUnit, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if ou == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/orgunits"
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, createdOrgUnit, err := motmedelHttpUtils.FetchJsonWithBody[*org_unit.OrgUnit](ctx, urlString, ou, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdOrgUnit, nil
}

// GetOrgUnit retrieves an organizational unit identified by orgUnitPath (e.g. "/Engineering/Frontend") or unique ID (e.g. "id:03ph8a2z1xdnme9").
func (c *Client) GetOrgUnit(ctx context.Context, customer string, orgUnitPath string, options ...fetch_config.Option) (*org_unit.OrgUnit, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}
	if orgUnitPath == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("org unit path"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	// Slashes in the org unit path are segment separators and must not be escaped.
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/orgunits/" + strings.TrimPrefix(orgUnitPath, "/")
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, ou, err := motmedelHttpUtils.FetchJson[*org_unit.OrgUnit](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return ou, nil
}

// UpdateOrgUnit updates an organizational unit identified by orgUnitPath or unique ID.
func (c *Client) UpdateOrgUnit(ctx context.Context, customer string, orgUnitPath string, ou *org_unit.OrgUnit, options ...fetch_config.Option) (*org_unit.OrgUnit, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}
	if orgUnitPath == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("org unit path"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if ou == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	// Slashes in the org unit path are segment separators and must not be escaped.
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/orgunits/" + strings.TrimPrefix(orgUnitPath, "/")
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPut))
	_, updatedOrgUnit, err := motmedelHttpUtils.FetchJsonWithBody[*org_unit.OrgUnit](ctx, urlString, ou, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updatedOrgUnit, nil
}

// DeleteOrgUnit deletes an organizational unit identified by orgUnitPath or unique ID.
func (c *Client) DeleteOrgUnit(ctx context.Context, customer string, orgUnitPath string, options ...fetch_config.Option) error {
	if customer == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}
	if orgUnitPath == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("org unit path"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	// Slashes in the org unit path are segment separators and must not be escaped.
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/orgunits/" + strings.TrimPrefix(orgUnitPath, "/")
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

type listOrgUnitsResponse struct {
	OrganizationUnits []*org_unit.OrgUnit `json:"organizationUnits"`
}

// ListOrgUnits retrieves all organizational units for the given customer ID (use "my_customer" for the authenticated account).
func (c *Client) ListOrgUnits(ctx context.Context, customer string, options ...fetch_config.Option) ([]*org_unit.OrgUnit, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/orgunits"

	query := url.Values{}
	query.Set("type", "all")
	urlObj.RawQuery = query.Encode()
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, resp, err := motmedelHttpUtils.FetchJson[*listOrgUnitsResponse](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	if resp == nil {
		return nil, nil
	}

	return resp.OrganizationUnits, nil
}

// Role operations

type listRolesResponse struct {
	Items         []*role.Role `json:"items"`
	NextPageToken string       `json:"nextPageToken"`
}

// ListRoles retrieves all roles for the given customer ID (use "my_customer" for the authenticated account).
func (c *Client) ListRoles(ctx context.Context, customer string, options ...fetch_config.Option) ([]*role.Role, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	var allRoles []*role.Role
	pageToken := ""

	for {
		urlObj := *c.baseUrl
		urlObj.Path += "customer/" + url.PathEscape(customer) + "/roles"

		query := url.Values{}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		urlObj.RawQuery = query.Encode()
		urlString := urlObj.String()

		paginatedOptions := append(c.config.FetchOptions, options...)
		_, resp, err := motmedelHttpUtils.FetchJson[*listRolesResponse](ctx, urlString, paginatedOptions...)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
		}

		if resp != nil {
			allRoles = append(allRoles, resp.Items...)

			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}

	return allRoles, nil
}

// CreateRole creates a new role for the given customer ID.
func (c *Client) CreateRole(ctx context.Context, customer string, r *role.Role, options ...fetch_config.Option) (*role.Role, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if r == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/roles"
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, createdRole, err := motmedelHttpUtils.FetchJsonWithBody[*role.Role](ctx, urlString, r, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdRole, nil
}

// GetRole retrieves a role identified by roleId.
func (c *Client) GetRole(ctx context.Context, customer string, roleId string, options ...fetch_config.Option) (*role.Role, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}
	if roleId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("role id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/roles/" + url.PathEscape(roleId)
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, r, err := motmedelHttpUtils.FetchJson[*role.Role](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return r, nil
}

// UpdateRole updates a role identified by roleId.
func (c *Client) UpdateRole(ctx context.Context, customer string, roleId string, r *role.Role, options ...fetch_config.Option) (*role.Role, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}
	if roleId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("role id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if r == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/roles/" + url.PathEscape(roleId)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPut))
	_, updatedRole, err := motmedelHttpUtils.FetchJsonWithBody[*role.Role](ctx, urlString, r, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updatedRole, nil
}

// DeleteRole deletes a role identified by roleId.
func (c *Client) DeleteRole(ctx context.Context, customer string, roleId string, options ...fetch_config.Option) error {
	if customer == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}
	if roleId == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("role id"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/roles/" + url.PathEscape(roleId)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

type listPrivilegesResponse struct {
	Items []*privilege.Privilege `json:"items"`
}

// ListPrivileges retrieves the privileges supported for building custom roles for the given customer ID.
func (c *Client) ListPrivileges(ctx context.Context, customer string, options ...fetch_config.Option) ([]*privilege.Privilege, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/roles/ALL/privileges"
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, resp, err := motmedelHttpUtils.FetchJson[*listPrivilegesResponse](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	if resp == nil {
		return nil, nil
	}

	return resp.Items, nil
}

// Role assignment operations

type listRoleAssignmentsResponse struct {
	Items         []*role_assignment.RoleAssignment `json:"items"`
	NextPageToken string                            `json:"nextPageToken"`
}

// ListRoleAssignments retrieves all role assignments for the given customer ID, optionally filtered by user key or role ID.
func (c *Client) ListRoleAssignments(ctx context.Context, customer string, options ...list_role_assignments_config.Option) ([]*role_assignment.RoleAssignment, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	listRoleAssignmentsConfig := list_role_assignments_config.New(options...)

	var allRoleAssignments []*role_assignment.RoleAssignment
	pageToken := ""

	for {
		urlObj := *c.baseUrl
		urlObj.Path += "customer/" + url.PathEscape(customer) + "/roleassignments"

		query := url.Values{}
		if listRoleAssignmentsConfig.UserKey != "" {
			query.Set("userKey", listRoleAssignmentsConfig.UserKey)
		}
		if listRoleAssignmentsConfig.RoleId != "" {
			query.Set("roleId", listRoleAssignmentsConfig.RoleId)
		}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		urlObj.RawQuery = query.Encode()
		urlString := urlObj.String()

		fetchOptions := append(c.config.FetchOptions, listRoleAssignmentsConfig.FetchOptions...)
		_, resp, err := motmedelHttpUtils.FetchJson[*listRoleAssignmentsResponse](ctx, urlString, fetchOptions...)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
		}

		if resp != nil {
			allRoleAssignments = append(allRoleAssignments, resp.Items...)

			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}

	return allRoleAssignments, nil
}

// CreateRoleAssignment creates a new role assignment for the given customer ID.
func (c *Client) CreateRoleAssignment(ctx context.Context, customer string, ra *role_assignment.RoleAssignment, options ...fetch_config.Option) (*role_assignment.RoleAssignment, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if ra == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/roleassignments"
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, createdRoleAssignment, err := motmedelHttpUtils.FetchJsonWithBody[*role_assignment.RoleAssignment](ctx, urlString, ra, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdRoleAssignment, nil
}

// GetRoleAssignment retrieves a role assignment identified by roleAssignmentId.
func (c *Client) GetRoleAssignment(ctx context.Context, customer string, roleAssignmentId string, options ...fetch_config.Option) (*role_assignment.RoleAssignment, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}
	if roleAssignmentId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("role assignment id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/roleassignments/" + url.PathEscape(roleAssignmentId)
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, ra, err := motmedelHttpUtils.FetchJson[*role_assignment.RoleAssignment](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return ra, nil
}

// DeleteRoleAssignment deletes a role assignment identified by roleAssignmentId.
func (c *Client) DeleteRoleAssignment(ctx context.Context, customer string, roleAssignmentId string, options ...fetch_config.Option) error {
	if customer == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}
	if roleAssignmentId == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("role assignment id"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "customer/" + url.PathEscape(customer) + "/roleassignments/" + url.PathEscape(roleAssignmentId)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

// Token operations

type listTokensResponse struct {
	Items []*token.Token `json:"items"`
}

// ListTokens retrieves the OAuth tokens issued to third-party applications for the user identified by userKey.
func (c *Client) ListTokens(ctx context.Context, userKey string, options ...fetch_config.Option) ([]*token.Token, error) {
	if userKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey) + "/tokens"
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, resp, err := motmedelHttpUtils.FetchJson[*listTokensResponse](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	if resp == nil {
		return nil, nil
	}

	return resp.Items, nil
}

// GetToken retrieves the OAuth token issued to the third-party application identified by clientId for the user identified by userKey.
func (c *Client) GetToken(ctx context.Context, userKey string, clientId string, options ...fetch_config.Option) (*token.Token, error) {
	if userKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}
	if clientId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("client id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey) + "/tokens/" + url.PathEscape(clientId)
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, t, err := motmedelHttpUtils.FetchJson[*token.Token](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return t, nil
}

// DeleteToken revokes the OAuth token issued to the third-party application identified by clientId for the user identified by userKey.
func (c *Client) DeleteToken(ctx context.Context, userKey string, clientId string, options ...fetch_config.Option) error {
	if userKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}
	if clientId == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("client id"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey) + "/tokens/" + url.PathEscape(clientId)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

// Application-specific password operations

type listAspsResponse struct {
	Items []*asp.Asp `json:"items"`
}

// ListAsps retrieves the application-specific passwords issued for the user identified by userKey.
func (c *Client) ListAsps(ctx context.Context, userKey string, options ...fetch_config.Option) ([]*asp.Asp, error) {
	if userKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey) + "/asps"
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, resp, err := motmedelHttpUtils.FetchJson[*listAspsResponse](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	if resp == nil {
		return nil, nil
	}

	return resp.Items, nil
}

// GetAsp retrieves the application-specific password identified by codeId for the user identified by userKey.
func (c *Client) GetAsp(ctx context.Context, userKey string, codeId int, options ...fetch_config.Option) (*asp.Asp, error) {
	if userKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey) + "/asps/" + strconv.Itoa(codeId)
	urlString := urlObj.String()

	options = append(c.config.FetchOptions, options...)
	_, a, err := motmedelHttpUtils.FetchJson[*asp.Asp](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return a, nil
}

// DeleteAsp revokes the application-specific password identified by codeId for the user identified by userKey.
func (c *Client) DeleteAsp(ctx context.Context, userKey string, codeId int, options ...fetch_config.Option) error {
	if userKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey) + "/asps/" + strconv.Itoa(codeId)
	urlString := urlObj.String()

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}
