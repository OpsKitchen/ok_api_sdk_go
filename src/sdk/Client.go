package sdk

import (
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
)

type Client struct {
	HttpClient     *http.Client
	Logger         *log.Logger
	RequestBuilder *RequestBuilder
}

func NewClient() *Client {
	return &Client {
		HttpClient: http.DefaultClient,
		RequestBuilder: &RequestBuilder {
			Config: NewConfig(),
			Credential: &Credential {},
		},
	}
}

func (client *Client) CallApi(api string, version string, params interface{}) (*ApiResult, error)  {
	var request *http.Request
	var response *http.Response
	var apiResult *ApiResult

	request = client.RequestBuilder.Build(api, version, params)
	response, err := client.HttpClient.Do(request)
	if err != nil {
		//client.Logger.Fatal("Do http request failed: " + api + " " + version)
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		//client.Logger.Fatal("Read response body failed: " + api + " " + version)
		return nil, err
	}

	err = json.Unmarshal(body, &apiResult)
	if err != nil {
		//client.Logger.Fatal("Json decode failed: " + api + " " + version)
		return nil, err
	}
	return apiResult, nil
}