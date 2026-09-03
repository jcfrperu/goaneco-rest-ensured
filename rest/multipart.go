package rest

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

// MultiPart represents a form-data part.
type MultiPart struct {
	ControlName string
	FileName    string
	ContentType string
	Content     []byte
	FilePath    string
}

// copyFileContent opens filePath, copies its content into dst, and closes the file.
// Using a helper prevents defer f.Close() from accumulating inside a loop.
func copyFileContent(dst io.Writer, filePath, controlName string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening file %q: %w", filePath, err)
	}
	defer f.Close()
	if _, err := io.Copy(dst, f); err != nil {
		return fmt.Errorf("writing file part %q: %w", controlName, err)
	}
	return nil
}

// escapeMIMEParam escapes double quotes and backslashes inside a MIME header parameter value
// so that Content-Disposition headers are not broken by special characters in field names or filenames.
func escapeMIMEParam(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r == '"' || r == '\\' {
			sb.WriteByte('\\')
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// buildMultipartBody creates a multipart/form-data body buffer and content-type header with boundary.
func buildMultipartBody(multiparts []MultiPart, formParams map[string][]string) ([]byte, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add form fields
	for name, values := range formParams {
		for _, val := range values {
			if err := writer.WriteField(name, val); err != nil {
				return nil, "", fmt.Errorf("writing form field %q: %w", name, err)
			}
		}
	}

	// Add multipart parts
	for _, part := range multiparts {
		filename := part.FileName

		h := make(textproto.MIMEHeader)
		if part.FilePath != "" && filename == "" {
			filename = filepath.Base(part.FilePath)
		}
		if filename != "" {
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeMIMEParam(part.ControlName), escapeMIMEParam(filename)))
		} else {
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, escapeMIMEParam(part.ControlName)))
		}
		ct := part.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		h.Set("Content-Type", ct)

		partWriter, err := writer.CreatePart(h)
		if err != nil {
			return nil, "", fmt.Errorf("creating multipart part %q: %w", part.ControlName, err)
		}

		if part.FilePath != "" {
			if err := copyFileContent(partWriter, part.FilePath, part.ControlName); err != nil {
				return nil, "", err
			}
		} else {
			if _, err := io.Copy(partWriter, bytes.NewReader(part.Content)); err != nil {
				return nil, "", fmt.Errorf("writing part data %q: %w", part.ControlName, err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("closing multipart writer: %w", err)
	}

	return body.Bytes(), writer.FormDataContentType(), nil
}
