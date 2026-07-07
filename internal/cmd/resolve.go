package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func defaultWeekStart() string {
	now := time.Now()
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	monday := now.AddDate(0, 0, -daysSinceMonday)
	return monday.Format("2006-01-02")
}

func weekStartOrDefault(flag string) string {
	if strings.TrimSpace(flag) != "" {
		return strings.TrimSpace(flag)
	}
	return defaultWeekStart()
}

func requireConfirm(confirm bool, action string) error {
	if !confirm {
		return fmt.Errorf("%s requires --confirm", action)
	}
	return nil
}

func resolvePersonID(personID, email string) (string, error) {
	personID = strings.TrimSpace(personID)
	email = strings.TrimSpace(email)
	if personID != "" {
		return personID, nil
	}
	if email == "" {
		return "", fmt.Errorf("one of --person-id or --email is required")
	}
	c, err := resolveClient(true)
	if err != nil {
		return "", err
	}
	resp, err := c.Get("/api/v1/persons", nil)
	if err != nil {
		handleAPIError(err)
	}
	var persons []map[string]any
	if err := json.Unmarshal(resp.Body, &persons); err != nil {
		return "", fmt.Errorf("parse persons list: %w", err)
	}
	want := strings.ToLower(email)
	for _, p := range persons {
		em, _ := p["email"].(string)
		if strings.ToLower(em) == want {
			id, _ := p["id"].(string)
			if id != "" {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("no person found with email %q", email)
}

func resolveProjectID(projectID, name, code string) (string, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID != "" {
		return projectID, nil
	}
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	if name == "" && code == "" {
		return "", fmt.Errorf("one of project ID, --name, or --code is required")
	}

	c, err := resolveClient(true)
	if err != nil {
		return "", err
	}
	resp, err := c.Get("/api/v1/projects", nil)
	if err != nil {
		handleAPIError(err)
	}
	var projects []map[string]any
	if err := json.Unmarshal(resp.Body, &projects); err != nil {
		return "", fmt.Errorf("parse projects list: %w", err)
	}

	if code != "" {
		want := strings.ToLower(code)
		for _, p := range projects {
			cn, _ := p["canonicalName"].(string)
			if strings.ToLower(cn) == want {
				id, _ := p["id"].(string)
				if id != "" {
					return id, nil
				}
			}
		}
		return "", fmt.Errorf("no project found with code/name %q", code)
	}

	want := strings.ToLower(name)
	var exact []string
	var partial []string
	for _, p := range projects {
		cn, _ := p["canonicalName"].(string)
		id, _ := p["id"].(string)
		if id == "" {
			continue
		}
		lcn := strings.ToLower(cn)
		if lcn == want {
			exact = append(exact, id)
		} else if strings.Contains(lcn, want) {
			partial = append(partial, id)
		}
	}
	switch {
	case len(exact) == 1:
		return exact[0], nil
	case len(exact) > 1:
		return "", fmt.Errorf("multiple projects match name %q; use --code or project ID", name)
	case len(partial) == 1:
		return partial[0], nil
	case len(partial) > 1:
		return "", fmt.Errorf("multiple projects match name %q; use --code or project ID", name)
	default:
		return "", fmt.Errorf("no project found with name %q", name)
	}
}

func readJSONFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("file must contain valid JSON")
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	return m, nil
}

func mergePayload(file string, fields map[string]any) ([]byte, error) {
	payload := map[string]any{}
	if file != "" {
		m, err := readJSONFile(file)
		if err != nil {
			return nil, err
		}
		payload = m
	}
	for k, v := range fields {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case string:
			if val != "" {
				payload[k] = val
			}
		case bool:
			payload[k] = val
		case []string:
			if len(val) > 0 {
				payload[k] = val
			}
		default:
			payload[k] = val
		}
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("no payload: use --file or field flags")
	}
	return json.Marshal(payload)
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
