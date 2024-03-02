package utils

import (
	"os/exec"
)

func PDFToText(src string, dest string) error {
	pdftotextArgs := []string{"-layout", src, dest}
	cmd := exec.Command("/usr/bin/pdftotext", pdftotextArgs...)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
