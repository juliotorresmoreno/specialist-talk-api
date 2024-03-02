package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func GenerateRandomFileName(prefix, suffix string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatal("Error generating random file name:", err)
	}
	return prefix + fmt.Sprintf("%x", b) + suffix
}

func ReadPDF(fileBuff string) (string, error) {
	attachment, err := ParseBase64File(fileBuff)
	if err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(attachment)
	if err != nil {
		return "", err
	}

	fileName := GenerateRandomFileName("attachment_", ".pdf")
	filePath := filepath.Join(os.TempDir(), fileName)

	err = os.WriteFile(filePath, decoded, 0644)
	if err != nil {
		return "", err
	}

	outputName := GenerateRandomFileName("attachment_", ".txt")
	outputPath := filepath.Join(os.TempDir(), outputName)

	err = PDFToText(filePath, outputPath)
	if err != nil {
		return "", err
	}

	foutput, err := os.Open(outputPath)
	if err != nil {
		return "", err
	}
	defer foutput.Close()

	output, err := io.ReadAll(foutput)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
