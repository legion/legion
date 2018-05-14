package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/legion/master/backends"
	"github.com/legion/master/spark"
	"github.com/legion/master/utils"
)

type Server struct {
	Router *mux.Router
}

var server Server
var httpPort = "7090"

var sparkClient spark.Client
var workersToExpire []backends.Worker
var azCLI *utils.CliExecutor
var backendProvider = ""
var isBusy = false

func Execute() {
	server = Server{
		Router: mux.NewRouter(),
	}

	server.tryProviderLogin()
	server.initEndpoints()
	server.watchSparkMasterLogs()

	http.Handle("/", server.Router)
	http.ListenAndServe("0.0.0.0:"+httpPort, server.Router)
}

func Status(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"status:": "running"})
}

func (s *Server) tryProviderLogin() {
	azureAppID := os.Getenv("AZURE_APP_ID")
	azureAppSecret := os.Getenv("AZURE_APP_SECRET")
	azureTenantID := os.Getenv("AZURE_TENANT_ID")

	if azureAppID != "" && azureAppSecret != "" && azureTenantID != "" {
		backendProvider = "aci"
		azCLI = utils.NewCliExecutor(azureAppID, azureAppSecret, azureTenantID)
		azCLI.Login()
	}
}

func (s *Server) watchSparkMasterLogs() {
	ticker := time.NewTicker(time.Second * 3)

	go func() {
		for range ticker.C {
			hostname, err := os.Hostname()
			apps, err := sparkClient.GetApplicationLogs(hostname)

			if err != nil {
				deleteCurrentWorkers()
			} else {
				if isBusy == false && len(apps) > 0 {
					isBusy = true
				}

				for _, app := range apps {
					for _, attempt := range app.Attempts {
						if attempt.Completed {
							deleteCurrentWorkers()
							break
						}
					}
				}
			}
		}
	}()
}

func deleteCurrentWorkers() {
	if isBusy {
		isBusy = false

		if len(workersToExpire) > 0 {
			for _, worker := range workersToExpire {
				if backendProvider == "aci" {
					err := azCLI.DeleteACI(worker.WorkerInfo["name"], worker.WorkerInfo["resourceGroup"])

					if err != nil {
						fmt.Println("Failed to delete worker: " + err.Error())
					} else {
						fmt.Println("Deleted worker " + worker.WorkerInfo["name"])
					}
				}
			}

			workersToExpire = []backends.Worker{}
		}
	}
}

func Submit(w http.ResponseWriter, r *http.Request) {
	if isBusy {
		respondWithError(w, http.StatusBadRequest, "server is busy running a spark application")
		return
	}
	file, handler, err := r.FormFile("file")

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var buf bytes.Buffer
	io.Copy(&buf, file)

	dataBytes := buf.Bytes()
	filePath := os.Getenv("HOME") + "/" + handler.Filename
	err = ioutil.WriteFile(filePath, dataBytes, 0644)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	cmdArgs := []string{}
	for key, values := range r.PostForm {
		if len(values) > 0 {
			val := values[0]
			cmdArgs = append(cmdArgs, key, val)
		}
	}

	blockManagerPort := os.Getenv("SPARK_BLOCKMANAGER_PORT")
	driverPort := os.Getenv("SPARK_DRIVER_PORT")
	hostname, _ := os.Hostname()
	cmdStr := "/spark/bin/spark-submit " + strings.Join(cmdArgs, " ") + " --master spark://" + hostname + ":7077 " + "--conf spark.driver.blockManager.port=" + blockManagerPort + " --conf spark.driver.port=" + driverPort + " --conf spark.driver.bindAddress=0.0.0.0 " + filePath

	cmd := exec.Command("sh", "-c", cmdStr)

	if err = cmd.Start(); err != nil {
		fmt.Println("Failed to start command: " + err.Error())
		respondWithError(w, 400, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status:": "ok"})
}

func ExpireWorkers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var workers []backends.Worker

	err := decoder.Decode(&workers)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	workersToExpire = append(workersToExpire, workers...)
	respondWithJSON(w, http.StatusOK, map[string]string{"status:": "ok"})
}

func (s *Server) initEndpoints() {
	s.Router.HandleFunc("/status", Status).Methods("GET")
	s.Router.HandleFunc("/submit", Submit).Methods("POST")
	s.Router.HandleFunc("/workers/expire", ExpireWorkers).Methods("POST")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(404)
	w.Write([]byte(`{"error": ` + message + `}`))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.Encode(payload)

	bytes := buffer.Bytes()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(bytes)
}
