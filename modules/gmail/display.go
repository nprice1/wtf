package gmail

import (
	b64 "encoding/base64"
	"fmt"
	"os"
	"os/exec"

	gmail "google.golang.org/api/gmail/v1"
)

func (widget *Widget) display() {
	message := widget.currentMessage()

	if message == nil {
		widget.View.SetText("No messages")
		return
	}

	str := fmt.Sprintf(
		"%s - \n From: %s\n Subject: [%s]%s\n\n %s",
		message.date,
		message.from,
		"white",
		message.subject,
		widget.getMessageContent(message.payload, message.payload.MimeType, message.id),
	)

	widget.View.SetText(str)
	widget.View.ScrollToBeginning()
}

func (widget *Widget) formatAllParts(rootPart *gmail.MessagePart, mimeType string) string {
	if rootPart.Filename != "" {
		return fmt.Sprintf("Attachment: %s", rootPart.Filename)
	}
	expectedMimeType := "text/html"
	if mimeType == "text/plain" {
		expectedMimeType = "text/plain"
	}
	if rootPart.Body.Data != "" && rootPart.MimeType == expectedMimeType {
		data, err := b64.URLEncoding.DecodeString(rootPart.Body.Data)
		if err != nil {
			return fmt.Sprintf("Error parsing data: %v", err)
		}
		return string(data)
	} else if rootPart.Parts != nil {
		str := ""
		for _, part := range rootPart.Parts {
			str = str + widget.formatAllParts(part, mimeType) + "\n"
		}
		return str
	}
	return ""
}

func (widget *Widget) getMessageContent(rootPart *gmail.MessagePart, mimeType string, messageId string) string {
	content := widget.formatAllParts(rootPart, mimeType)
	filePath := widget.getMessageDir() + messageId + ".html"
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		f, err := os.Create(filePath)
		if err != nil {
			return fmt.Sprintf("Error creating file: %v", err)
		}
		err = f.Chmod(775)
		if err != nil {
			return fmt.Sprintf("Error modifying file: %v", err)
		}
		_, err = f.WriteString(content)
		if err != nil {
			return fmt.Sprintf("Error writing to file: %v", err)
		}
	}
	out, err := exec.Command("/usr/local/bin/w3m", "-dump", "-cols", "200", filePath).Output()
	if err != nil {
		return fmt.Sprintf("Error with w3m command file: %v", err)
	}
	return string(out)
}
