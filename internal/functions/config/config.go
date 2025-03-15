package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
)

const fmd os.FileMode = 0o666

func LoadConfigServer(
	pth string,
) (*apimodels.CfgServer, error) {
	file, err := os.OpenFile(pth, os.O_RDONLY|os.O_EXCL, fmd)
	if err != nil {
		return nil, fmt.Errorf("LoadConfigServer->OF: %w", err)
	}

	defer file.Close()

	params := &apimodels.CfgServer{}

	byteValue, _ := io.ReadAll(file)

	err = json.Unmarshal(byteValue, &params)
	if err != nil {
		return nil, fmt.Errorf("LoadConfigServer->Unma: %w", err)
	}

	return params, nil
}

func LoadConfigAgent(
	pth string,
) (*apimodels.CfgAgent, error) {
	file, err := os.OpenFile(pth, os.O_RDONLY|os.O_EXCL, fmd)
	if err != nil {
		return nil, fmt.Errorf("LoadConfigAgent->OF: %w", err)
	}

	defer file.Close()

	params := &apimodels.CfgAgent{}

	byteValue, _ := io.ReadAll(file)

	err = json.Unmarshal(byteValue, &params)
	if err != nil {
		return nil, fmt.Errorf("LoadConfigAgent->Unma: %w", err)
	}

	return params, nil
}
