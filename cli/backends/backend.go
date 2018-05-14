package backends

type Backend interface {
	CreateMaster(cores int, memory string, location string) (Master, error)
	Submit(executorCores int, executorMemory string, numExecutors int, driverMemory string, driverCores int, location string, master string) (SubmitResponse, error)
	Name() string
}

type Master struct {
	IPAddress string
	FQDN      string
	ExtraInfo map[string]string
}

type Worker struct {
	WorkerInfo map[string]string `json:"workerInfo`
}

type SubmitResponse struct {
	MasterInfo  Master
	WorkersInfo []Worker
}
