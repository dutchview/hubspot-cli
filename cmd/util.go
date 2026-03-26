package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func printJSON(v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func printRawJSON(data json.RawMessage) {
	var v interface{}
	if json.Unmarshal(data, &v) == nil {
		printJSON(v)
	} else {
		fmt.Println(string(data))
	}
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func formatTimestamp(ts string) string {
	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, ts); err == nil {
			return t.Local().Format("2006-01-02 15:04")
		}
	}
	return ts
}

func prop(props map[string]interface{}, key string) string {
	if v, ok := props[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func confirmAction(prompt string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

type crmObject struct {
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
}

type crmListResponse struct {
	Results []crmObject `json:"results"`
	Paging  *struct {
		Next *struct {
			After string `json:"after"`
		} `json:"next"`
	} `json:"paging"`
}

func parseCRMList(data json.RawMessage) (*crmListResponse, error) {
	var resp crmListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func parseCRMObject(data json.RawMessage) (*crmObject, error) {
	var obj crmObject
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
