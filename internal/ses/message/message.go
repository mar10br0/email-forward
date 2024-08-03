package sesmessage

import (
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/mar10br0/email-forward/internal/log"
)

func parseHeader(messageData string) (header http.Header) {
	header = make(http.Header)
	lines := strings.Split(strings.Split(messageData, "\r\n\r\n")[0], "\r\n")
	var headers []string
	for _, line := range lines {
		if len(headers) > 0 && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			headers[len(headers)-1] += strings.Trim(line, " \t")
		} else {
			headers = append(headers, line)
		}
	}
	for _, pair := range headers {
		key := strings.SplitN(pair, ":", 2)[0]
		values := strings.Split(strings.TrimPrefix(pair, key+": "), ",")
		for _, value := range values {
			header.Add(key, value)
		}
	}
	return
}

func Fetch(svc *s3.S3, recipient string, messageID string) (message string, err error) {
	domain := strings.Split(recipient, "@")[1]
	user := strings.Split(recipient, "@")[0]
	bucket := domain + "-inbox"
	key := user + "/" + messageID
	var msgObj *s3.GetObjectOutput
	if msgObj, err = svc.GetObject(&s3.GetObjectInput{Bucket: &bucket, Key: &key}); err != nil {
		return
	}
	defer msgObj.Body.Close()

	msgBuf := new(strings.Builder)
	var copied int64
	if copied, err = io.Copy(msgBuf, msgObj.Body); err != nil {
		return
	}
	log.Debug("Fetched %s", log.Plurality(copied, "bytes"))
	message = msgBuf.String()
	log.Response(parseHeader(message))

	return
}

func Delete(svc *s3.S3, recipient string, messageID string) (err error) {
	domain := strings.Split(recipient, "@")[1]
	user := strings.Split(recipient, "@")[0]
	bucket := domain + "-inbox"
	key := user + "/" + messageID
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: &bucket, Key: &key})
	return
}
