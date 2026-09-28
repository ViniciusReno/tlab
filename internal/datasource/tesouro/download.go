package tesouro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ViniciusReno/tlab/internal/bond"
)

const PackageURL = "https://www.tesourotransparente.gov.br/ckan/api/3/action/package_show?id=taxas-dos-titulos-ofertados-pelo-tesouro-direto"
const packageID = "df56aa42-484a-4a59-8184-7676580c81e3"
const maxMetadataBytes = 1 << 20

// Client permits replacing the HTTP transport in offline tests. Endpoints, limits,
// timeouts, and redirect policy remain fixed; no user URL is accepted.
type Client struct{ Transport http.RoundTripper }

type Download struct {
	Dataset
	ResourceID string
	SourceURL  string
}

func officialURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Host == "www.tesourotransparente.gov.br" && u.User == nil && u.Fragment == "" && u.Opaque == ""
}

func (c Client) Fetch(ctx context.Context) (Download, error) {
	client := &http.Client{Transport: c.Transport, Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 || !officialURL(req.URL.String()) {
				return bond.SourceUnavailable
			}
			return nil
		}}
	metadata, _, err := download(ctx, client, PackageURL, "application/json", maxMetadataBytes)
	if err != nil {
		return Download{}, fmt.Errorf("CKAN metadata download: %w", err)
	}
	var envelope struct {
		Success bool `json:"success"`
		Result  struct {
			ID        string `json:"id"`
			State     string `json:"state"`
			Private   *bool  `json:"private"`
			Resources []struct {
				ID        string `json:"id"`
				PackageID string `json:"package_id"`
				State     string `json:"state"`
				Format    string `json:"format"`
				URL       string `json:"url"`
			} `json:"resources"`
		} `json:"result"`
	}
	if err := json.Unmarshal(metadata, &envelope); err != nil || !envelope.Success || envelope.Result.ID != packageID || envelope.Result.State != "active" || envelope.Result.Private == nil || *envelope.Result.Private {
		return Download{}, fmt.Errorf("CKAN response must identify the active public official package: %w", bond.SchemaChanged)
	}
	var result Download
	count := 0
	for _, resource := range envelope.Result.Resources {
		if resource.State != "active" || !strings.EqualFold(resource.Format, "CSV") {
			continue
		}
		count++
		if resource.ID == "" || resource.PackageID != packageID || !officialURL(resource.URL) {
			return Download{}, fmt.Errorf("CSV resource identity or official HTTPS URL is invalid: %w", bond.SchemaChanged)
		}
		result.ResourceID, result.SourceURL = resource.ID, resource.URL
	}
	if count != 1 {
		return Download{}, fmt.Errorf("expected exactly one active CSV resource, found %d: %w", count, bond.SchemaChanged)
	}
	body, finalURL, err := download(ctx, client, result.SourceURL, "text/csv", MaxDatasetBytes)
	if err != nil {
		return result, fmt.Errorf("official CSV download: %w", err)
	}
	result.SourceURL = finalURL
	result.Dataset, err = ParseDataset(bytes.NewReader(body), finalURL)
	if err != nil {
		return result, fmt.Errorf("official CSV has missing columns, malformed or duplicate rows, or no data: %w", err)
	}
	if len(result.Quotes) == 0 {
		return result, fmt.Errorf("no supported Prefixado rows: %w", bond.MissingQuote)
	}
	return result, nil
}

func download(ctx context.Context, client *http.Client, source, contentType string, limit int64) ([]byte, string, error) {
	if !officialURL(source) {
		return nil, "", bond.SourceUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, "", bond.SourceUnavailable
	}
	req.Header.Set("Accept", contentType)
	req.Header.Set("User-Agent", "Tesouro-Lab/1 (local educational data sync)")
	response, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed, timed out, was canceled, or used a disallowed redirect: %w", bond.SourceUnavailable)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP status %d: %w", response.StatusCode, bond.SourceUnavailable)
	}
	actualType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || actualType != contentType {
		return nil, "", fmt.Errorf("expected HTTP Content-Type %s: %w", contentType, bond.SchemaChanged)
	}
	if response.ContentLength > limit {
		return nil, "", fmt.Errorf("download exceeds the %d-byte limit: %w", limit, bond.InvalidInput)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, "", fmt.Errorf("download was interrupted or incomplete: %w", bond.SourceUnavailable)
	}
	if len(body) == 0 || int64(len(body)) > limit {
		return nil, "", fmt.Errorf("download is empty or exceeds the %d-byte limit: %w", limit, bond.InvalidInput)
	}
	return body, response.Request.URL.String(), nil
}
