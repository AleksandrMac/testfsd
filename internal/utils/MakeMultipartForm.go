package utils

import (
	"bytes"
	"mime/multipart"
)

// MultypartFormData make multypart/formdata and boundary for facts
func MultypartFormData(data map[string]string) (*bytes.Buffer, string, error) {
	buf := &bytes.Buffer{}

	writer := multipart.NewWriter(buf)
	defer writer.Close()

	for key, val := range data {
		if err := writer.WriteField(key, val); err != nil {
			return nil, "", err
		}
	}
	return buf, writer.Boundary(), nil
}
