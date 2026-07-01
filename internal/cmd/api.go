package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	apiQueries []string
	apiBody    string
)

var apiCmd = &cobra.Command{
	Use:   "api METHOD PATH",
	Short: "Generic API request",
	Long: `Send an arbitrary HTTP request to the API.

Examples:
  tt api GET /api/v1/persons
  tt api PUT /api/v1/persons/UUID/manager --body '{"managerId":"..."}'
  tt api POST /api/v1/persons/import --query onDuplicate=update --body @payload.json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		method := strings.ToUpper(args[0])
		path := args[1]

		query, err := parseQueries(apiQueries)
		if err != nil {
			return err
		}

		body, err := readBody(apiBody)
		if err != nil {
			return err
		}

		requireAuth := path != "/health"
		c, err := resolveClient(requireAuth)
		if err != nil {
			return err
		}

		resp, err := c.Do(method, path, query, body)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

func init() {
	apiCmd.Flags().StringArrayVar(&apiQueries, "query", nil, "query parameter key=value (repeatable)")
	apiCmd.Flags().StringVar(&apiBody, "body", "", "JSON body or @file.json")
}

func parseQueries(pairs []string) (url.Values, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	v := url.Values{}
	for _, pair := range pairs {
		key, val, ok := strings.Cut(pair, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid query %q (use key=value)", pair)
		}
		v.Add(key, val)
	}
	return v, nil
}

func readBody(spec string) ([]byte, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}
	if strings.HasPrefix(spec, "@") {
		data, err := os.ReadFile(strings.TrimPrefix(spec, "@"))
		if err != nil {
			return nil, fmt.Errorf("read body file: %w", err)
		}
		return data, nil
	}
	if !json.Valid([]byte(spec)) {
		return nil, fmt.Errorf("body must be valid JSON or @file")
	}
	return []byte(spec), nil
}
