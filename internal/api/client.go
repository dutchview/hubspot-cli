package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const baseURL = "https://api.hubapi.com"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	reqURL := baseURL + endpoint
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == 204 {
		return nil, nil
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Message)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// --- Tickets ---

func (c *Client) GetTicket(ticketID string, properties []string, associations []string) (json.RawMessage, error) {
	params := url.Values{}
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	if len(associations) > 0 {
		params.Set("associations", strings.Join(associations, ","))
	}
	endpoint := "/crm/v3/objects/tickets/" + ticketID
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	return c.doRequest("GET", endpoint, nil)
}

func (c *Client) ListTickets(limit int, properties []string, after string) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	if after != "" {
		params.Set("after", after)
	}
	return c.doRequest("GET", "/crm/v3/objects/tickets?"+params.Encode(), nil)
}

func (c *Client) SearchTickets(filterGroups []map[string]interface{}, properties []string, limit int, after int, sorts []map[string]string) (json.RawMessage, error) {
	body := map[string]interface{}{
		"filterGroups": filterGroups,
		"properties":   properties,
		"limit":        limit,
	}
	if after > 0 {
		body["after"] = after
	}
	if len(sorts) > 0 {
		body["sorts"] = sorts
	}
	return c.doRequest("POST", "/crm/v3/objects/tickets/search", body)
}

func (c *Client) CreateTicket(properties map[string]string) (json.RawMessage, error) {
	body := map[string]interface{}{
		"properties": properties,
	}
	return c.doRequest("POST", "/crm/v3/objects/tickets", body)
}

func (c *Client) UpdateTicket(ticketID string, properties map[string]string) (json.RawMessage, error) {
	body := map[string]interface{}{
		"properties": properties,
	}
	return c.doRequest("PATCH", "/crm/v3/objects/tickets/"+ticketID, body)
}

func (c *Client) DeleteTicket(ticketID string) error {
	_, err := c.doRequest("DELETE", "/crm/v3/objects/tickets/"+ticketID, nil)
	return err
}

// --- Conversations ---

func (c *Client) GetThread(threadID string) (json.RawMessage, error) {
	return c.doRequest("GET", "/conversations/v3/conversations/threads/"+threadID, nil)
}

func (c *Client) GetThreadMessages(threadID string) (json.RawMessage, error) {
	return c.doRequest("GET", "/conversations/v3/conversations/threads/"+threadID+"/messages", nil)
}

func (c *Client) AddComment(threadID string, text string, richText string) (json.RawMessage, error) {
	body := map[string]interface{}{
		"type": "COMMENT",
		"text": text,
	}
	if richText != "" {
		body["richText"] = richText
	}
	return c.doRequest("POST", "/conversations/v3/conversations/threads/"+threadID+"/messages", body)
}

func (c *Client) SendMessage(threadID string, text string, richText string, senderActorID string, channelID string, channelAccountID string, recipients []map[string]string) (json.RawMessage, error) {
	body := map[string]interface{}{
		"type":             "MESSAGE",
		"text":             text,
		"senderActorId":    senderActorID,
		"channelId":        channelID,
		"channelAccountId": channelAccountID,
		"recipients":       recipients,
	}
	if richText != "" {
		body["richText"] = richText
	}
	return c.doRequest("POST", "/conversations/v3/conversations/threads/"+threadID+"/messages", body)
}

// --- Contacts ---

func (c *Client) GetContact(contactID string, properties []string) (json.RawMessage, error) {
	params := url.Values{}
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	endpoint := "/crm/v3/objects/contacts/" + contactID
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	return c.doRequest("GET", endpoint, nil)
}

func (c *Client) ListContacts(limit int, properties []string, after string) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	if after != "" {
		params.Set("after", after)
	}
	return c.doRequest("GET", "/crm/v3/objects/contacts?"+params.Encode(), nil)
}

func (c *Client) SearchContacts(filterGroups []map[string]interface{}, properties []string, limit int) (json.RawMessage, error) {
	body := map[string]interface{}{
		"filterGroups": filterGroups,
		"properties":   properties,
		"limit":        limit,
	}
	return c.doRequest("POST", "/crm/v3/objects/contacts/search", body)
}

// --- Companies ---

func (c *Client) GetCompany(companyID string, properties []string) (json.RawMessage, error) {
	params := url.Values{}
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	endpoint := "/crm/v3/objects/companies/" + companyID
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	return c.doRequest("GET", endpoint, nil)
}

func (c *Client) ListCompanies(limit int, properties []string, after string) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	if after != "" {
		params.Set("after", after)
	}
	return c.doRequest("GET", "/crm/v3/objects/companies?"+params.Encode(), nil)
}

func (c *Client) SearchCompanies(filterGroups []map[string]interface{}, properties []string, limit int) (json.RawMessage, error) {
	body := map[string]interface{}{
		"filterGroups": filterGroups,
		"properties":   properties,
		"limit":        limit,
	}
	return c.doRequest("POST", "/crm/v3/objects/companies/search", body)
}

// --- Deals ---

func (c *Client) GetDeal(dealID string, properties []string) (json.RawMessage, error) {
	params := url.Values{}
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	endpoint := "/crm/v3/objects/deals/" + dealID
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	return c.doRequest("GET", endpoint, nil)
}

func (c *Client) ListDeals(limit int, properties []string, after string) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	if after != "" {
		params.Set("after", after)
	}
	return c.doRequest("GET", "/crm/v3/objects/deals?"+params.Encode(), nil)
}

func (c *Client) SearchDeals(filterGroups []map[string]interface{}, properties []string, limit int) (json.RawMessage, error) {
	body := map[string]interface{}{
		"filterGroups": filterGroups,
		"properties":   properties,
		"limit":        limit,
	}
	return c.doRequest("POST", "/crm/v3/objects/deals/search", body)
}

// --- Owners ---

func (c *Client) GetOwner(ownerID string) (json.RawMessage, error) {
	return c.doRequest("GET", "/crm/v3/owners/"+ownerID, nil)
}

func (c *Client) ListOwners(limit int, after string) (json.RawMessage, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if after != "" {
		params.Set("after", after)
	}
	endpoint := "/crm/v3/owners"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	return c.doRequest("GET", endpoint, nil)
}

// --- Pipelines ---

func (c *Client) GetPipelines(objectType string) (json.RawMessage, error) {
	return c.doRequest("GET", "/crm/v3/pipelines/"+objectType, nil)
}

func (c *Client) GetPipelineStages(objectType, pipelineID string) (json.RawMessage, error) {
	return c.doRequest("GET", "/crm/v3/pipelines/"+objectType+"/"+pipelineID+"/stages", nil)
}

// --- Associations ---

func (c *Client) GetAssociations(objectType, objectID, toObjectType string) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("/crm/v4/objects/%s/%s/associations/%s", objectType, objectID, toObjectType)
	return c.doRequest("GET", endpoint, nil)
}

// --- Engagement / Activity Timeline ---

func (c *Client) SearchEngagements(objectType string, filterGroups []map[string]interface{}, properties []string, limit int) (json.RawMessage, error) {
	body := map[string]interface{}{
		"filterGroups": filterGroups,
		"properties":   properties,
		"limit":        limit,
	}
	return c.doRequest("POST", "/crm/v3/objects/"+objectType+"/search", body)
}
