package compresspdf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/InspectorGadget/adobe-go-sdk/auth"
)

type AdobeAssetResponse struct {
	UploadUri string `json:"uploadUri"`
	AssetID   string `json:"assetID"`
}

type AdobeCompressResponse struct {
	Location string `json:"location"`
	AssetID  string `json:"assetID"`
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
	Status  string                      `json:"status"`
	Asset   AdobeJobStatusAssetResponse `json:"asset"`
	AssetID string                      `json:"assetID"`
	Error   struct {
		Status  int    `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

var (
	ErrNotAuthenticated = errors.New("not authenticated")
)

func CreateNewAsset(auth *auth.AdobeAuthenticationContext) (AdobeAssetResponse, error) {
	if auth.GetAccessToken() == "" {
		return AdobeAssetResponse{}, ErrNotAuthenticated
	}

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	data := map[string]string{
		"mediaType": "application/pdf",
	}
	jsonPayload, err := json.Marshal(data)
	if err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("failed to marshal json data: %w", err)
	}

	request, err := http.NewRequest("POST", "https://pdf-services.adobe.io/assets", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("failed to create new request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+auth.GetAccessToken())
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", auth.GetClientID())

	response, err := client.Do(request)
	if err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var jsonData AdobeAssetResponse
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return AdobeAssetResponse{}, fmt.Errorf("failed to unmarshal json data: %w", err)
	}

	return jsonData, nil
}

func CompressPDF(auth *auth.AdobeAuthenticationContext, fileData []byte) (AdobeJobStatusResponse, error) {
	if auth.GetAccessToken() == "" {
		return AdobeJobStatusResponse{}, ErrNotAuthenticated
	}

	assetService, err := CreateNewAsset(auth)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create new asset: %w", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	request, err := http.NewRequest("PUT", assetService.UploadUri, bytes.NewBuffer(fileData))
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create new request: %w", err)
	}
	request.Header.Set("Content-Type", "application/pdf")

	response, err := httpClient.Do(request)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to upload asset: %s, Body: %s", response.Status, string(body))
	}

	result, err := CreateCompressionJob(auth, assetService.AssetID)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create compression job: %w", err)
	}

	return result, nil
}

func CreateCompressionJob(auth *auth.AdobeAuthenticationContext, assetId string) (AdobeJobStatusResponse, error) {
	if auth.GetAccessToken() == "" {
		return AdobeJobStatusResponse{}, ErrNotAuthenticated
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	data := map[string]any{
		"assetID":          assetId,
		"compressionLevel": "LOW",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to marshal json data: %w", err)
	}

	request, err := http.NewRequest("POST", "https://pdf-services.adobe.io/operation/compresspdf", bytes.NewBuffer(jsonData))
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create new request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+auth.GetAccessToken())
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", auth.GetClientID())

	response, err := httpClient.Do(request)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to compress PDF: %s, Body: %s, AssetID: %s", response.Status, string(body), assetId)
	}

	compressResponse := AdobeCompressResponse{
		Location: response.Header.Get("Location"),
		AssetID:  assetId,
	}

	value, err := GetJobStatus(auth, compressResponse)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to get job status: %w", err)
	}

	return value, nil
}

func GetJobStatus(auth *auth.AdobeAuthenticationContext, compressResponse AdobeCompressResponse) (AdobeJobStatusResponse, error) {
	if auth.GetAccessToken() == "" {
		return AdobeJobStatusResponse{}, ErrNotAuthenticated
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	request, err := http.NewRequest("GET", compressResponse.Location, nil)
	if err != nil {
		return AdobeJobStatusResponse{}, fmt.Errorf("failed to create new request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+auth.GetAccessToken())
	request.Header.Set("X-API-Key", auth.GetClientID())

	for range 30 { // Retry for a certain number of times. Default: 30
		response, err := httpClient.Do(request)
		if err != nil {
			return AdobeJobStatusResponse{}, fmt.Errorf("failed to execute request: %w", err)
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return AdobeJobStatusResponse{}, fmt.Errorf("failed to get job status: %s", response.Status)
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return AdobeJobStatusResponse{}, fmt.Errorf("failed to read response body: %w", err)
		}

		var jobResponse AdobeJobStatusResponse
		if err := json.Unmarshal(body, &jobResponse); err != nil {
			return AdobeJobStatusResponse{}, fmt.Errorf("failed to unmarshal json data: %w", err)
		}

		switch jobResponse.Status {
		case "done":
			return jobResponse, nil
		case "failed":
			return AdobeJobStatusResponse{}, fmt.Errorf("compression failed: %s", jobResponse.Error.Message)
		case "in progress":
			log.Println("[Compress PDF] Still polling:", jobResponse.Status)
			time.Sleep(2 * time.Second) // Add a delay to prevent excessive polling
			continue
		default:
			return AdobeJobStatusResponse{}, fmt.Errorf("unknown job status: %s", jobResponse.Status)
		}

	}

	return AdobeJobStatusResponse{}, fmt.Errorf("polling timed out")
}
