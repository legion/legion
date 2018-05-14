package backends

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"

	"github.com/legion/cli/utils"
)

type AzureContainerInstancesBackend struct {
	CLI *utils.CliExecutor
}

type AzureStorageAccountCredentials struct {
	AccountName string
	AccountKey  string
}

type ACIResources struct {
	Master  utils.ACI
	Workers []utils.ACI
}

func (a *AzureContainerInstancesBackend) getMasterFQDN(name string, location string) string {
	return name + "." + location + ".azurecontainer.io"
}

func (a *AzureContainerInstancesBackend) GenerateMasterName() string {
	return "legionmaster" + strconv.Itoa(utils.RangeIn(1000, 9999))
}

func (a *AzureContainerInstancesBackend) getACIResources(executorCores int, executorMemory float64, numExecutors int, driverMemory float64, driverCores int, resourceGroup string, location string) *ACIResources {
	masterName := a.GenerateMasterName()
	fqdn := a.getMasterFQDN(masterName, location)
	storageCreds := a.GetStorageCredentials()

	resources := &ACIResources{
		Master: utils.ACI{
			CPU:           driverCores,
			Memory:        driverMemory,
			Name:          masterName,
			Location:      location,
			ResourceGroup: resourceGroup,
			Image:         MasterDockerImage,
			EnvVars: map[string]string{"ENABLE_INIT_DAEMON": "false", "DNS_LABEL": fqdn, "SPARK_PUBLIC_DNS": fqdn, "SPARK_BLOCKMANAGER_PORT": "3500", "SPARK_DRIVER_PORT": "3000", "AZURE_STORAGE_ACCOUNT_NAME": storageCreds.AccountName, "AZURE_STORAGE_ACCOUNT_KEY": storageCreds.AccountKey,
				"AZURE_APP_ID": os.Getenv("APP_ID"), "AZURE_APP_SECRET": os.Getenv("PASSWORD"), "AZURE_TENANT_ID": os.Getenv("TENANT_ID")},
			PublicAddress: true,
			DNSLabel:      masterName,
			Ports:         []string{"7090", "8080", "3000", "3500", "7077"},
		},
	}

	for i := 0; i < numExecutors; i++ {
		resources.Workers = append(resources.Workers, utils.ACI{
			CPU:           executorCores,
			Memory:        executorMemory,
			Name:          "legionworker" + strconv.Itoa(utils.RangeIn(1000, 9999)),
			Location:      location,
			ResourceGroup: resourceGroup,
			Image:         WorkerDockerImage,
			PublicAddress: true,
			EnvVars:       map[string]string{"ENABLE_INIT_DAEMON": "false", "SPARK_WORKER_PORT": "5000", "AZURE_STORAGE_ACCOUNT_NAME": storageCreds.AccountName, "AZURE_STORAGE_ACCOUNT_KEY": storageCreds.AccountKey},
			Ports:         []string{"5000"},
		})
	}

	return resources
}

func (a *AzureContainerInstancesBackend) GetStorageCredentials() AzureStorageAccountCredentials {
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")

	return AzureStorageAccountCredentials{
		AccountName: accountName,
		AccountKey:  accountKey,
	}
}

func (a *AzureContainerInstancesBackend) generateResourceGroupName() string {
	return "legion-" + strconv.Itoa(rand.Int())
}

func (a *AzureContainerInstancesBackend) Submit(executorCores int, executorMemory string, numExecutors int, driverMemory string, driverCores int, location string, existingMaster string) (SubmitResponse, error) {
	executorMemory = strings.Replace(strings.ToLower(executorMemory), "g", "", -1)
	driverMemory = strings.Replace(strings.ToLower(driverMemory), "g", "", -1)

	executorMemoryFloat, _ := strconv.ParseFloat(executorMemory, 64)
	driverMemoryFloat, _ := strconv.ParseFloat(driverMemory, 64)

	resourceGroup := a.generateResourceGroupName()

	resources := a.getACIResources(executorCores, executorMemoryFloat, numExecutors, driverMemoryFloat, driverCores, resourceGroup, location)

	a.CLI = utils.NewCliExecutor(os.Getenv("APP_ID"), os.Getenv("PASSWORD"), os.Getenv("TENANT_ID"))
	a.CLI.Login()

	if (existingMaster != "" && numExecutors > 0) || (existingMaster == "") {
		err := a.CLI.CreateResourceGroup(resourceGroup, location)

		if err != nil {
			return SubmitResponse{}, err
		}
	}

	masterACI := resources.Master
	workersACI := resources.Workers

	masterAddress := ""

	if existingMaster == "" {
		ip, err := a.deployMaster(masterACI)

		if err != nil {
			return SubmitResponse{}, errors.New("failed deploying Master - " + err.Error())
		}

		masterAddress = ip
	} else {
		masterAddress = existingMaster
	}

	err := a.deployWorkers(workersACI, masterAddress)

	if err != nil {
		fmt.Println(errors.New("failed deploying Workers - " + err.Error()))
		return SubmitResponse{}, err
	}

	response := a.GetSubmitResponse(*resources, masterAddress, location, resourceGroup)
	return response, nil
}

func (a *AzureContainerInstancesBackend) GetSubmitResponse(resources ACIResources, masterAddress string, location string, resourceGroup string) SubmitResponse {
	submitResponse := SubmitResponse{}

	submitResponse.MasterInfo = Master{
		IPAddress: masterAddress,
		FQDN:      a.getMasterFQDN(resources.Master.Name, location),
	}

	for _, workerACI := range resources.Workers {
		worker := Worker{}
		worker.WorkerInfo = map[string]string{}
		worker.WorkerInfo["name"] = workerACI.Name
		worker.WorkerInfo["resourceGroup"] = resourceGroup

		submitResponse.WorkersInfo = append(submitResponse.WorkersInfo, worker)
	}

	return submitResponse
}

func (a *AzureContainerInstancesBackend) CreateMaster(cores int, memory string, location string) (Master, error) {
	memory = strings.Replace(strings.ToLower(memory), "g", "", -1)
	memoryFloat, _ := strconv.ParseFloat(memory, 64)
	resourceGroup := a.generateResourceGroupName()
	resources := a.getACIResources(0, 0, 0, memoryFloat, cores, resourceGroup, location)
	master := resources.Master

	a.CLI = utils.NewCliExecutor(os.Getenv("APP_ID"), os.Getenv("PASSWORD"), os.Getenv("TENANT_ID"))
	a.CLI.Login()

	err := a.CLI.CreateResourceGroup(resourceGroup, location)

	if err != nil {
		return Master{}, err
	}

	ip, err := a.deployMaster(master)
	if err != nil {
		return Master{}, err
	}

	fqdn := a.getMasterFQDN(master.Name, location)
	return Master{IPAddress: ip, FQDN: fqdn}, nil
}

func (a *AzureContainerInstancesBackend) deployMaster(master utils.ACI) (string, error) {
	fmt.Println("deploying Spark Master")
	_, err := a.CLI.CreateACI(master)

	if err != nil {
		return "", err
	}

	fmt.Println("master deployed - ~50 seconds to Running state")
	running := false
	ipAddress := ""

	stopCount := 0
	stopThreshold := 40

	for running == false {
		aciJSON, err := a.CLI.GetACI(master.Name, master.ResourceGroup)

		if err != nil {
			return "", err
		}

		state, _ := jsonparser.GetString([]byte(aciJSON), "instanceView", "state")
		fmt.Println("master state: " + state)

		if state == "Running" {
			running = true
			address, _ := jsonparser.GetString([]byte(aciJSON), "ipAddress", "ip")
			ipAddress = address
		} else {
			stopCount++
			if stopThreshold == stopCount {
				return "", errors.New("deployment failed. Please try again")
			}
		}

		time.Sleep(5 * time.Second)
	}

	return ipAddress, nil
}

func (a *AzureContainerInstancesBackend) deployWorkers(workers []utils.ACI, masterIP string) error {
	numberOfWorkers := len(workers)

	if numberOfWorkers > 0 {
		fmt.Println("deploying Spark Workers")

		wg := sync.WaitGroup{}
		wg.Add(numberOfWorkers)

		for _, worker := range workers {
			go a.deployWorker(masterIP, worker, &wg)
		}

		wg.Wait()
		fmt.Println("Finished deploying " + strconv.Itoa(len(workers)) + " workers. ~40 seconds for each worker to join the cluster.")
	}

	return nil
}

func (a *AzureContainerInstancesBackend) deployWorker(masterIP string, worker utils.ACI, wg *sync.WaitGroup) {
	worker.EnvVars["SPARK_MASTER"] = masterIP + ":7077"

	_, err := a.CLI.CreateACI(worker)

	if err != nil {
		fmt.Println("Error deploying worker: " + err.Error())
		wg.Done()
		return
	}

	fmt.Println("worker " + worker.Name + " deployed")
	wg.Done()
}

func (a *AzureContainerInstancesBackend) Name() string {
	return "aci"
}

func NewAzureContainerInstancesBackend() *AzureContainerInstancesBackend {
	return &AzureContainerInstancesBackend{}
}
