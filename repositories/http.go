package repositories

import (
	"PaaS-deployment-sqs-trigger/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func DeployApp(
	appId string,
) (
	r *models.DeployAppResponseModel,
	e error,
) {

	var err error

	apiURL := "http://ec2-34-226-197-212.compute-1.amazonaws.com:8080/deploy"

	requestData := map[string]string{
		"app_identifier": appId,
	}

	requestDataJSON, err := json.Marshal(requestData)
	if err != nil {
		return r, fmt.Errorf("new error happened trying to build http request data %s", err.Error())
	}

	log.Println("endpoint", apiURL)
	log.Println("method", http.MethodPost)
	log.Println("requestData", string(requestDataJSON))

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(string(requestDataJSON)))
	if err != nil {
		return r, fmt.Errorf("new error happened trying to build http request %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return r, fmt.Errorf("new error trying to make http request %s", err.Error())
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)

	if err != nil {
		return r, fmt.Errorf("new error trying to make http request %s", err.Error())
	}

	log.Println("responseBody", string(data))

	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("an occured error trying to unmashal response %s", err.Error())
	}

	// Check the response status code
	if res.StatusCode != http.StatusCreated {
		return r, fmt.Errorf("request failed with status code: %d. Data: %s", res.StatusCode, string(data))
	}

	if r == nil {
		return r, fmt.Errorf("request failed with status code: %d. Data: %s. Data is null", res.StatusCode, string(data))
	}

	return r, nil
}
