package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

func CheckNudity(filename string) {
	url := "https://api.sightengine.com/1.0/check.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	// Pass image URL instead of file
	_ = writer.WriteField("media", "https://feignedly-unpaired-amiya.ngrok-free.dev/uploads/products/"+filename)
	_ = writer.WriteField("models", "nudity-2.1")
	_ = writer.WriteField("api_user", "1350960651")
	_ = writer.WriteField("api_secret", "xD7trXQ3EDEzJsd4Msy5bZzVZCXADoJf")
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
