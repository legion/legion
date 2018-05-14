package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type CliExecutor struct {
	AppID    string
	TenantID string
	Password string
	AksName  string
	LoggedIn bool
}

type ACI struct {
	CPU           int
	Memory        float64
	ResourceGroup string
	Location      string
	Name          string
	Image         string
	EnvVars       map[string]string
	PublicAddress bool
	DNSLabel      string
	Ports         []string
}

func GetAllowedACICores() []int {
	return []int{1, 2, 3, 4}
}

func GetAllowedACIMemory() []float64 {
	mem := []float64{}

	min := 0.5
	max := 13.5

	for i := min; i <= max; i += 0.5 {
		mem = append(mem, i+0.5)
	}

	return mem
}

func NewCliExecutor(appID string, password string, tenantID string) *CliExecutor {
	if appID == "" {
		panic("appId cannot be empty")
	} else if password == "" {
		panic("password cannot be empty")
	} else if tenantID == "" {
		panic("tenantId cannot be empty")
	} else {
		return &CliExecutor{AppID: appID, Password: password, TenantID: tenantID}
	}
}

func (c *CliExecutor) CreateResourceGroup(name string, location string) error {
	if !c.LoggedIn {
		return errors.New("must be logged in before executing commands")
	}

	cmd := "az"
	cmdArgs := []string{"group", "create", "-n", name, "-l", location}
	_, err := exec.Command(cmd, cmdArgs...).Output()

	if err != nil {
		return err
	}

	return nil
}

func (c *CliExecutor) DeleteACI(name string, resourceGroup string) error {
	cmd := "az"
	cmdArgs := []string{"container", "delete", "-n", name, "-g", resourceGroup, "-y"}

	fmt.Println("--------- CMD: " + strings.Join(cmdArgs, " "))

	_, err := exec.Command(cmd, cmdArgs...).Output()

	if err != nil {
		return err
	}

	return nil
}

func (c *CliExecutor) GetACI(name string, resourceGroup string) (string, error) {
	cmd := "az"
	cmdArgs := []string{"container", "show", "-n", name, "-g", resourceGroup}

	out, err := exec.Command(cmd, cmdArgs...).Output()

	if err != nil {
		return "", err
	}

	return string(out[:]), nil
}

func (c *CliExecutor) CreateACI(aci ACI) (string, error) {
	if !c.LoggedIn {
		return "", errors.New("must be logged in before executing commands")
	}

	envvars := []string{}

	for k, v := range aci.EnvVars {
		envvars = append(envvars, k+"="+v)
	}

	cmd := "az"
	cmdArgs := []string{"container", "create", "-n", aci.Name, "-g", aci.ResourceGroup, "-l", aci.Location, "--image", aci.Image, "--cpu", strconv.Itoa(aci.CPU), "--memory", strconv.FormatFloat(aci.Memory, 'f', 0, 64)}

	if len(aci.EnvVars) > 0 {
		cmdArgs = append(cmdArgs, "-e")

		for k, v := range aci.EnvVars {
			cmdArgs = append(cmdArgs, k+"="+v)
		}
	}

	if aci.PublicAddress {
		cmdArgs = append(cmdArgs, "--ip-address", "public")
	}

	if aci.DNSLabel != "" {
		cmdArgs = append(cmdArgs, "--dns-name-label", aci.DNSLabel)
	}

	if len(aci.Ports) > 0 {
		cmdArgs = append(cmdArgs, "--ports")

		for _, port := range aci.Ports {
			cmdArgs = append(cmdArgs, port)
		}
	}

	out, err := exec.Command(cmd, cmdArgs...).Output()

	if err != nil {
		return "", err
	}

	return string(out[:]), nil
}

func (c *CliExecutor) Login() {
	if !c.LoggedIn {
		cmd := "az"
		cmdArgs := []string{"login", "--service-principal", "-u", c.AppID, "-p", c.Password, "-t", c.TenantID}
		if err := exec.Command(cmd, cmdArgs...).Run(); err != nil {
			panic("az login failed")
		}

		fmt.Println("az login successful")
		c.LoggedIn = true
	}
}
