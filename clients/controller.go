package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/sidecar/config"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Controller TODO
type Controller interface {
	Register() error
	GetNGINXConfig(version *time.Time) (string, error)
	GetCredentials() (TenantCredentials, error)
}

// Registry TODO
type Registry struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// Kafka TODO
type Kafka struct {
	APIKey   string   `json:"api_key"`
	AdminURL string   `json:"admin_url"`
	RestURL  string   `json:"rest_url"`
	Brokers  []string `json:"brokers"`
	User     string   `json:"user"`
	Password string   `json:"password"`
	SASL     bool     `json:"sasl"`
}

type tenantInfo struct {
	ID          string            `json:"id"`
	Credentials TenantCredentials `json:"credentials"`
	Port        int               `json:"port"`
}

// TenantCredentials credentials
type TenantCredentials struct {
	Kafka    Kafka    `json:"kafka"`
	Registry Registry `json:"registry"`
}

type controller struct {
	config *config.Config
	client http.Client
}

// NewController TODO
func NewController(conf *config.Config) Controller {
	return &controller{
		config: conf,
		client: http.Client{},
	}
}

// Register TODO
func (c *controller) Register() error {

	bodyJSON := tenantInfo{
		ID: c.config.Tenant.ID,
		Credentials: TenantCredentials{
			Kafka: Kafka{
				APIKey:   c.config.Kafka.APIKey,
				User:     c.config.Kafka.Username,
				Password: c.config.Kafka.Password,
				Brokers:  c.config.Kafka.Brokers,
				RestURL:  c.config.Kafka.RestURL,
				AdminURL: c.config.Kafka.RestURL,
			},
			Registry: Registry{
				URL:   c.config.Registry.URL,
				Token: c.config.Registry.Token,
			},
		},
		Port: c.config.Nginx.Port,
	}

	bodyBytes, err := json.Marshal(bodyJSON)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"url":    c.config.Controller.URL + "/v1/tenants",
			"method": "POST",
		}).Warn("Error marshalling JSON body")
		return err
	}
	reader := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest("POST", c.config.Controller.URL+"/v1/tenants", reader)
	req.Header.Set("Content-type", "application/json")
	// TODO set Authorization header
	req.Header.Set("Authorization", c.config.Tenant.Token)

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusConflict {
			respBytes, _ := ioutil.ReadAll(resp.Body)
			// TODO return custom error object
			return fmt.Errorf("ID already present: %v", string(respBytes))
		}
		return fmt.Errorf("Controller error.  Expected %v got status code %v.  %v", http.StatusCreated, resp.StatusCode, string(respBytes))
	}

	return nil
}

func (c *controller) GetNGINXConfig(version *time.Time) (string, error) {

	url, err := url.Parse(c.config.Controller.URL + "/v1/tenants/" + c.config.Tenant.ID + "/nginx")
	if err != nil {
		return "", err
	}
	if version != nil {
		query := url.Query()
		query.Add("version", version.Format(time.RFC3339))
		url.RawQuery = query.Encode()
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
			"tenant_id": c.config.Tenant.ID,
		}).Warn("Error building request to get rules from controller")
		return "", err
	}
	//TODO set auth header
	req.Header.Set("Authorization", c.config.Tenant.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
			"tenant_id": c.config.Tenant.ID,
		}).Warn("Failed to retrieve rules from controller")
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		logrus.Info("No new rules received")
		return "", nil

	} else if resp.StatusCode != http.StatusOK {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			//			"request_id": reqID,
			"tenant_id": c.config.Tenant.ID,
			"body":      string(respBytes),
		}).Warn("Controller returned bad response code")
		return "", errors.New("Controller returned bad response code") // FIXME: custom error?
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
			"tenant_id": c.config.Tenant.ID,
		}).Warn("Error reading rules JSON from controller")
		return "", err
	}

	return string(body), err
}

func (c *controller) GetCredentials() (TenantCredentials, error) {

	respJSON := struct {
		Credentials TenantCredentials `json:"credentials"`
	}{}

	url, err := url.Parse(c.config.Controller.URL + "/v1/tenants/" + c.config.Tenant.ID)
	if err != nil {
		return respJSON.Credentials, err
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
			"tenant_id": c.config.Tenant.ID,
		}).Warn("Error building request to get creds from Controller")
		return respJSON.Credentials, err
	}
	//TODO set auth header
	req.Header.Set("Authorization", c.config.Tenant.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
			"tenant_id": c.config.Tenant.ID,
		}).Warn("Failed to retrieve creds from Controller")
		return respJSON.Credentials, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			//			"request_id": reqID,
			"tenant_id": c.config.Tenant.ID,
			"body":      string(respBytes),
		}).Warn("Controller returned bad response code")
		if resp.StatusCode == http.StatusNotFound {
			respBytes, _ := ioutil.ReadAll(resp.Body)
			// TODO return custom error object
			return respJSON.Credentials, fmt.Errorf("ID not found: %v", string(respBytes))
		}
		return respJSON.Credentials, errors.New("Controller returned bad response code") // FIXME: custom error?
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
			"tenant_id": c.config.Tenant.ID,
		}).Warn("Error reading rules JSON from Controller")
		return respJSON.Credentials, err
	}

	err = json.Unmarshal(body, &respJSON)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			//			"request_id": reqID,
			"tenant_id": c.config.Tenant.ID,
		}).Warn("Error reading creds JSON from Controller")
		return respJSON.Credentials, err
	}

	return respJSON.Credentials, nil
}
