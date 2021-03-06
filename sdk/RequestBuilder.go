package sdk

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk/model"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type RequestBuilder struct {
	Config     *model.Config
	Credential *model.Credential
}

func (rb *RequestBuilder) Build(api string, version string, params interface{}) (*http.Request, error) {
	paramJson, err := rb.getParamsJson(params)
	if err != nil {
		return nil, err
	}

	deviceId := rb.Credential.DeviceId
	if rb.Credential.DeviceId == "" {
		sdkDeviceId, err := rb.getDeviceId()
		if err != nil {
			return nil, err
		}
		deviceId = sdkDeviceId
	}

	timestamp := rb.getTimestamp()
	gatewayUrl := rb.getGatewayUrl()
	requestBody := rb.getPostBody(api, version, paramJson, timestamp)

	DefaultLogger.Debug("[API SDK] Gateway url: " + gatewayUrl)
	DefaultLogger.Debug("[API SDK] Api: " + api + " " + version)
	DefaultLogger.Debug("[API SDK] Timestamp: " + timestamp)
	DefaultLogger.Debug("[API SDK] Request param: " + paramJson)

	//init http request
	request, err := http.NewRequest(http.MethodPost, gatewayUrl, strings.NewReader(requestBody))
	if err != nil {
		errMsg := "sdk: can not create http request object: " + err.Error()
		DefaultLogger.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	//set headers
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set(rb.Config.AppKeyFieldName, rb.Credential.AppKey)
	request.Header.Set(rb.Config.AppMarketIdFieldName, rb.Config.AppMarketIdValue)
	request.Header.Set(rb.Config.AppVersionFieldName, rb.Config.AppVersionValue)
	request.Header.Set(rb.Config.DeviceIdFieldName, deviceId)
	request.Header.Set(rb.Config.SessionIdFieldName, rb.Credential.SessionId)
	request.Header.Set(rb.Config.SignFieldName, rb.getSign(deviceId, rb.Credential.SessionId, api, version, paramJson, timestamp))
	DefaultLogger.Debug("[API SDK] Request header:", request.Header)
	return request, nil
}

func (rb *RequestBuilder) getDeviceId() (string, error) {
	uuidFile := rb.Config.DeviceIdFilePath
	if _, err := os.Stat(uuidFile); err != nil {
		deviceId := uuid.NewV1().String()
		if err := ioutil.WriteFile(uuidFile, []byte(deviceId), 0644); err != nil {
			errMsg := "sdk: failed to write uuid file [" + uuidFile + "]: " + err.Error()
			DefaultLogger.Error(errMsg)
			return "", errors.New(errMsg)
		}
		return deviceId, nil
	} else {
		deviceId, err := ioutil.ReadFile(uuidFile)
		if err != nil {
			errMsg := "sdk: failed to read uuid file [" + uuidFile + "]: " + err.Error()
			DefaultLogger.Error(errMsg)
			return "", errors.New(errMsg)
		}
		return string(deviceId), nil
	}
}

func (rb *RequestBuilder) getGatewayUrl() string {
	urlObj := url.URL{
		Path: rb.Config.GatewayPath,
	}
	if rb.Config.DisableSSL {
		urlObj.Scheme = "http"
	} else {
		urlObj.Scheme = "https"
	}
	if rb.Config.GatewayPort != 0 { //port number configured
		urlObj.Host = fmt.Sprintf("%s:%s", rb.Config.GatewayHost, strconv.Itoa(rb.Config.GatewayPort))
	} else {
		urlObj.Host = rb.Config.GatewayHost
	}
	return urlObj.String()
}

func (rb *RequestBuilder) getParamsJson(params interface{}) (string, error) {
	if params == nil {
		return "null", nil
	}
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		errMsg := "sdk: can not encode api parameter as json: " + err.Error()
		DefaultLogger.Error(errMsg)
		return "", errors.New(errMsg)
	}
	return string(jsonBytes), nil
}

func (rb *RequestBuilder) getPostBody(api string, version string, paramJson string, timestamp string) string {
	values := &url.Values{}
	values.Add(rb.Config.ApiFieldName, api)
	values.Add(rb.Config.VersionFieldName, version)
	values.Add(rb.Config.TimestampFieldName, timestamp)
	values.Add(rb.Config.ParamsFieldName, paramJson)
	return values.Encode()
}

func (rb *RequestBuilder) getSign(deviceId, sessionId, api, version, paramJson, timestamp string) string {
	hashObj := md5.New()
	stringToBeSign := rb.Credential.Secret + deviceId + sessionId + api + version + paramJson + timestamp
	io.WriteString(hashObj, stringToBeSign)
	return fmt.Sprintf("%x", hashObj.Sum(nil))
}

//get timestamp in microsecond
func (rb *RequestBuilder) getTimestamp() string {
	return strconv.FormatInt(time.Now().UnixNano()/1e3, 10)
}
