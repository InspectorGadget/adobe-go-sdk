package exportpdf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/InspectorGadget/adobe-go-sdk/auth"
)

type AdobeAssetResponse struct {
	UploadUri string `json:"uploadUri"`
	AssetID   string `json:"assetID"`
}

type AdobeJobStatusAssetMetadataResponse struct {
	Type string `json:"type"`
	Size int    `json:"size"`
}

type AdobeJobStatusAssetResponse struct {
	Metadata    AdobeJobStatusAssetMetadataResponse `json:"metadata"`
	DownloadUri string                              `json:"downloadUri"`
}

type AdobeJobStatusResponse struct {
	Status string                      `json:"status"`
	Asset  AdobeJobStatusAssetResponse `json:"asset"`
	Error  struct {
		Status  int    `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

var (
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrUploadFailed     = errors.New("failed to upload asset")
	ErrExportFailed     = errors.New("failed to export PDF")
	ErrPollFailed       = errors.New("failed to poll for status")
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		DisableCompression: true,
	},
	Timeout: 10 * time.Second, // Set a reasonable timeout
}

func CreateNewAsset(auth *auth.AdobeAuthenticationContext, format string) (AdobeAssetResponse, error) {
	if auth.GetAccessToken() == "" {
		return AdobeAssetResponse{}, ErrNotAuthenticated
	}

	data := map[string]string{
		"mediaType": strings.ToLower(format),
	}
	jsonPayload, err := json.Marshal(data)
	if err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("failed to marshal json: %w", err)
	}

	request, err := http.NewRequest("POST", "https://pdf-services.adobe.io/assets", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+auth.GetAccessToken())
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", auth.GetClientID())

	response, err := httpClient.Do(request)
	if err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("http request failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return AdobeAssetResponse{}, fmt.Errorf("failed to create asset: %s, %s", response.Status, string(body))
	}

	var jsonData AdobeAssetResponse
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return jsonData, nil
}

func ExportPDF(auth *auth.AdobeAuthenticationContext, fileData []byte) (AdobeJobStatusResponse, error) {
	if auth.GetAccessToken() == "" {
		return AdobeJobStatusResponse{}, ErrNotAuthenticated
	}

	assetService, err := CreateNewAsset(auth, "application/pdf")
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create new asset: %w", err)
	}

	request, err := http.NewRequest("PUT", assetService.UploadUri, bytes.NewBuffer(fileData))
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/pdf")

	response, err := httpClient.Do(request)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("http request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return AdobeJobStatusResponse{}, fmt.Errorf("%w: %s, %s", ErrUploadFailed, response.Status, string(body))
	}

	// Export the PDF
	jobResponse, err := CreateExportJob(auth, assetService.AssetID)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create export job: %w", err)
	}

	return jobResponse, nil
}

func CreateExportJob(auth *auth.AdobeAuthenticationContext, assetId string) (AdobeJobStatusResponse, error) {
	if auth.GetAccessToken() == "" {
		return AdobeJobStatusResponse{}, ErrNotAuthenticated
	}

	data := map[string]any{
		"assetID":      assetId,
		"targetFormat": "docx",
		"ocrLang":      "en-US",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to marshal json: %w", err)
	}

	request, err := http.NewRequest("POST", "https://pdf-services.adobe.io/operation/exportpdf", bytes.NewBuffer(jsonData))
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+auth.GetAccessToken())
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", auth.GetClientID())

	response, err := httpClient.Do(request)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("http request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		return AdobeJobStatusResponse{}, fmt.Errorf("%w: %s, %s, %s", ErrExportFailed, response.Status, string(body), assetId)
	}

	// Poll for status
	pollingUrl := response.Header.Get("Location")
	if pollingUrl == "" {
		return AdobeJobStatusResponse{}, fmt.Errorf("%w: Location header is missing", ErrExportFailed)
	}

	jobResponse, err := PollForStatus(auth, pollingUrl)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to poll for status: %w", err)
	}

	return jobResponse, nil
}

func PollForStatus(auth *auth.AdobeAuthenticationContext, url string) (AdobeJobStatusResponse, error) {
	if auth.GetAccessToken() == "" {
		return AdobeJobStatusResponse{}, ErrNotAuthenticated
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+auth.GetAccessToken())
	request.Header.Set("X-API-Key", auth.GetClientID())

	for range 30 { // Retry for a certain number of times. Default: 30
		response, err := httpClient.Do(request)
		if err != nil {
			return AdobeJobStatusResponse{}, fmt.Errorf("http request failed: %w", err)
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return AdobeJobStatusResponse{}, fmt.Errorf("failed to read response body: %w", err)
		}

		if response.StatusCode != http.StatusOK {
			return AdobeJobStatusResponse{}, fmt.Errorf("%w: %s, %s", ErrPollFailed, response.Status, string(body))
		}

		var jobResponse AdobeJobStatusResponse
		err = json.Unmarshal(body, &jobResponse)
		if err != nil {
			return AdobeJobStatusResponse{}, fmt.Errorf("failed to unmarshal json: %w", err)
		}

		if jobResponse.Status == "in progress" {
			log.Println("[Export PDF] Still polling:", jobResponse.Status)
			time.Sleep(2 * time.Second) // Wait before polling again
			continue
		}

		if jobResponse.Status == "failed" {
			return AdobeJobStatusResponse{}, fmt.Errorf("export failed: %s", jobResponse.Error.Message)
		}

		return jobResponse, nil
	}

	return AdobeJobStatusResponse{}, fmt.Errorf("polling timed out")
}
