package misc

import (
	"errors"
	"io/ioutil"
	"net/http"

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
)

var controllerHost string

type Config struct {
	Controller string `yaml:"controller"`
}

func ReadConfig(file string) (Config, error) {
	var config Config
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(raw, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

type Params struct {
	Args map[string]string
}

func NewParams() Params {
	p := Params{}
	p.Args = map[string]string{}
	return p
}

func (p *Params) AddParam(key, value string) {
	p.Args[key] = value
}

func (p *Params) ToString() string {
	str := ""
	for k, v := range p.Args {
		str += (k + "=" + v + "&")
	}
	return str[:len(str)-1]
}

func Report(controllerURL, path string, params Params) error {
	arg := params.ToString()
	request, err := http.NewRequest("POST", "http://"+controllerURL+"/"+path+"?"+arg, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Connection", "Keep-Alive")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		logrus.WithField("response", string(body)).Info("reported to controller")
		return nil
	} else {
		logrus.WithField("response", string(body)).Info("failed reporting to controller")
		return errors.New("HTTP response code isn't 200")
	}
}
