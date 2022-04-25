package ServiceData

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type (
	DataRole struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	DataPermissionsItem struct {
		Type            string   `json:"type"`
		PermissionsList []string `json:"permissions_list"`
	}

	DataTaskStatusItem struct {
		Type         string   `json:"type"`
		StatusesList []string `json:"statuses_list"`
	}

	DataRolePermissionsItem struct {
		Role            string   `json:"role"`
		PermissionsList []string `json:"permissions_list"`
	}
)

//------------

type ServiceData struct {
	basePath     string
	data         Data
	localization Localization
}

func NewServiceData(basePath string) *ServiceData {
	return &ServiceData{
		basePath: basePath,
	}
}

//GetDataFromFile returns data from json files
///		TODO create tests
func (d *ServiceData) GetDataFromFile(fileName string, logger *logrus.Logger) (Data, error) {
	var data Data

	path := filepath.Join(d.basePath, fileName)
	fileData, err := os.ReadFile(path)
	if err != nil {
		return Data{}, err
	}

	if len(fileData) == 0 {
		return Data{}, fmt.Errorf("data file is empty")
	}
	if err := json.NewDecoder(bytes.NewBuffer(fileData)).Decode(&data); err != nil {
		return Data{}, err
	}

	return data, nil
}

func (d *ServiceData) GetLocalizationsFromFile() (Localization, error) {

	return Localization{}, errors.New("")
}

type Data struct {
	Roles           []DataRole                `json:"roles"`
	PermissionTypes []string                  `json:"permission_types"`
	PermissionList  []DataPermissionsItem     `json:"permission_list"`
	RolePermissions []DataRolePermissionsItem `json:"role_permissions"`
	TaskStatusTypes []string                  `json:"task_status_types"`
	TaskStatusList  []DataTaskStatusItem      `json:"task_status_list"`
}

type Localization struct {
	Lang string
}
