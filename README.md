# Adobe GO SDK

## Description
Adobe GO SDK is a Go client for Adobe's APIs. It provides a simple and easy-to-use interface for interacting with Adobe's services, including Adobe Creative Cloud, Adobe Document Cloud, and Adobe Experience Cloud.

## Features
- Export PDF to any format (Partially implemented)
- Compress PDF (Implemented)

## Installation
To install the Adobe GO SDK, use the following command:

```bash
go get github.com/InspectorGadget/adobe-go-sdk
```

## Usage
```go
package main

import (
    "fmt"
    "log"

    "github.com/InspectorGadget/adobe-go-sdk/auth"
)

// Initialize the Adobe SDK, and authenticate
func initiateAuthentication() error {
	authInput := auth.AdobeAuthInput{
		ClientID:     "<client_id>",
		ClientSecret: "<client_secret>",
	}
	response, err := authInput.Authenticate()
	if err != nil {
		return errors.New("please check your credentials")
	}

	return nil
}

func init() {
    err := initiateAuthentication()
    if err != nil {
        log.Fatalf("Failed to authenticate: %v", err)
    }
}

func main() {
    // Example usage of the SDK
    fmt.Println("Adobe GO SDK initialized and authenticated successfully.")

    // 1. Getting the authentication information
    fmt.println("Access Token: ", auth.AdobeSession.AccessToken)
    fmt.println("Expires In: ", auth.AdobeSession.ExpiresIn)
    fmt.println("Token Type: ", auth.AdobeSession.TokenType)

    // 2. Calling the Export PDF Function
    value, err := exportpdf.ExportPDF(*auth.AdobeSession, []byte("<your_file_buffer>"))
    if err != nil {
        log.Fatalf("Failed to export PDF: %v", err)
    }

    response := map[string]string{
        "downloadUri": value.Asset.DownloadUri,
        "fileSize":    value.Asset.Metadata.Size,
        "fileType":    value.Asset.Metadata.Type,
    }
    fmt.println("Exported PDF response: ", response)
}
```

## Contributing
If you would like to contribute to the Adobe GO SDK, please fork the repository and submit a pull request. We welcome contributions of all kinds, including bug fixes, new features, and documentation improvements.

## Maintainers
- [InspectorGadget](https://github.com/InspectorGadget)
