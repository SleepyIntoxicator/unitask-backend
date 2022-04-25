package ServiceData

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestServiceData_GetDataFromFile(t *testing.T) {
	sd := NewServiceData("service")

	loadedData, err := sd.GetDataFromFile("service_roles.json", nil)
	if err != nil {
		t.Errorf("coudn't get data from file: %s", err)
		t.Fail()
		return
	}



	assert.NotNil(t, loadedData.Roles)
	assert.NotNil(t, loadedData.PermissionTypes)
	assert.NotNil(t, loadedData.PermissionList)
	assert.NotNil(t, loadedData.RolePermissions)
	assert.NotNil(t, loadedData.TaskStatusTypes)
	assert.NotNil(t, loadedData.TaskStatusList)

	if t.Failed() {
		return
	}

	fmt.Printf("Roles:\n")
	for i := range loadedData.Roles {
		fmt.Printf("%-25s \\ %q\n", loadedData.Roles[i].Name, loadedData.Roles[i].Description)
	}

	fmt.Printf("\n\n/////////////////////////////////////////////\n\n")

	fmt.Printf("PermissionTypes: %s", strings.Join(loadedData.PermissionTypes, ", "))

	for _, p := range loadedData.PermissionList {
		fmt.Printf("\n\n\tPermission type: %s\n\tPermissions:\t", p.Type)
		for i := range p.PermissionsList {
			fmt.Printf(" %s", p.PermissionsList[i])
		}
	}
	fmt.Printf("\n\n/////////////////////////////////////////////\n\n")

	fmt.Printf("Role permissions:")

	for _, r := range loadedData.RolePermissions {
		fmt.Printf("\n\n\tRole: %s\n\tPermissions:\t", r.Role)
		for i := range r.PermissionsList {
			fmt.Printf("\n\t\t\t%s ", r.PermissionsList[i])
		}
	}

	fmt.Printf("\n\n/////////////////////////////////////////////\n\n")

	fmt.Printf("TaskStatusTypes: %s", strings.Join(loadedData.TaskStatusTypes, ", "))

	for _, p := range loadedData.TaskStatusList {
		fmt.Printf("\n\n\tTaskStatus type: %s\n\tTaskStatuses:\t", p.Type)
		for i := range p.StatusesList {
			fmt.Printf("%s ", p.StatusesList[i])
		}
	}
}