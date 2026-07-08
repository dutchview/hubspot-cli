package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func (c *Client) SendMessage(threadID string, text string, richText string, senderActorID string, channelID string, channelAccountID string, recipients []map[string]interface{}) (json.RawMessage, error) {
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

// --- Notes ---

func (c *Client) GetNote(noteID string, properties []string) (json.RawMessage, error) {
	params := url.Values{}
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	endpoint := "/crm/v3/objects/notes/" + noteID
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	return c.doRequest("GET", endpoint, nil)
}

func (c *Client) ListNotes(limit int, properties []string, after string) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	if len(properties) > 0 {
		params.Set("properties", strings.Join(properties, ","))
	}
	if after != "" {
		params.Set("after", after)
	}
	return c.doRequest("GET", "/crm/v3/objects/notes?"+params.Encode(), nil)
}

func (c *Client) SearchNotes(filterGroups []map[string]interface{}, properties []string, limit int, after int, sorts []map[string]string) (json.RawMessage, error) {
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
	return c.doRequest("POST", "/crm/v3/objects/notes/search", body)
}

func (c *Client) CreateNote(body string, attachmentIDs []string) (json.RawMessage, error) {
	props := map[string]string{
		"hs_note_body": body,
		"hs_timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if len(attachmentIDs) > 0 {
		props["hs_attachment_ids"] = strings.Join(attachmentIDs, ";")
	}
	payload := map[string]interface{}{
		"properties": props,
	}
	return c.doRequest("POST", "/crm/v3/objects/notes", payload)
}

func (c *Client) UploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("copy file: %w", err)
	}

	optsPart, err := writer.CreateFormField("options")
	if err != nil {
		return "", fmt.Errorf("create options field: %w", err)
	}
	optsPart.Write([]byte(`{"access":"PRIVATE"}`))

	folderPart, err := writer.CreateFormField("folderPath")
	if err != nil {
		return "", fmt.Errorf("create folder field: %w", err)
	}
	folderPart.Write([]byte("/hubspot-cli-uploads"))

	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"/files/v3/files", &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	return result.ID, nil
}

func (c *Client) UpdateNote(noteID string, properties map[string]string) (json.RawMessage, error) {
	body := map[string]interface{}{
		"properties": properties,
	}
	return c.doRequest("PATCH", "/crm/v3/objects/notes/"+noteID, body)
}

func (c *Client) DeleteNote(noteID string) error {
	_, err := c.doRequest("DELETE", "/crm/v3/objects/notes/"+noteID, nil)
	return err
}

// --- Generic CRM Object CRUD ---

func (c *Client) CreateObject(objectType string, properties map[string]string) (json.RawMessage, error) {
	body := map[string]interface{}{
		"properties": properties,
	}
	return c.doRequest("POST", "/crm/v3/objects/"+objectType, body)
}

func (c *Client) UpdateObject(objectType, objectID string, properties map[string]string) (json.RawMessage, error) {
	body := map[string]interface{}{
		"properties": properties,
	}
	return c.doRequest("PATCH", "/crm/v3/objects/"+objectType+"/"+objectID, body)
}

func (c *Client) DeleteObject(objectType, objectID string) error {
	_, err := c.doRequest("DELETE", "/crm/v3/objects/"+objectType+"/"+objectID, nil)
	return err
}

// --- Associations ---

func (c *Client) GetAssociations(objectType, objectID, toObjectType string) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("/crm/v4/objects/%s/%s/associations/%s", objectType, objectID, toObjectType)
	return c.doRequest("GET", endpoint, nil)
}

func (c *Client) CreateAssociation(fromType, fromID, toType, toID string) error {
	// Look up the correct association type ID
	typeID, err := c.getAssociationTypeID(fromType, toType)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/crm/v4/objects/%s/%s/associations/%s/%s", fromType, fromID, toType, toID)
	body := []map[string]interface{}{
		{
			"associationCategory": "HUBSPOT_DEFINED",
			"associationTypeId":   typeID,
		},
	}
	_, err = c.doRequest("PUT", endpoint, body)
	return err
}

func (c *Client) getAssociationTypeID(fromType, toType string) (int, error) {
	endpoint := fmt.Sprintf("/crm/v4/associations/%s/%s/labels", fromType, toType)
	data, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}
	var resp struct {
		Results []struct {
			TypeID   int    `json:"typeId"`
			Category string `json:"category"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, err
	}
	for _, r := range resp.Results {
		if r.Category == "HUBSPOT_DEFINED" {
			return r.TypeID, nil
		}
	}
	return 0, fmt.Errorf("no association type found for %s -> %s", fromType, toType)
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
