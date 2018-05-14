# Legion: Serverless Spark For Cloud Native Platforms

Legion is a system for deploying Apache Spark on different serverless compute cloud platforms, allowing for on-demand, per second billable deployments of Spark jobs without any infrastructure to manage.

*Note: Legion is experimental software*

## About

Legion currently supports Azure Container Instances as a serverless backend platform.
Legion provides a simple CLI to submit jobs to backend platforms.
The legion submit command is similar to Spark's submit command, allowing for a dynamic creation of serverless components as the Spark executors.

## How it works

Legion is composed of two main components: the legion master and the legion worker.
The Legion master acts as the Spark Master and Driver, and also implements a lightweight Go based RESTful API server for submitting jobs to the Spark Master.

Legion uses Spark in Standalone mode.

For each executor requested, Legion will reach the backend provider which in turn decides what is the best way to represent the role of executor in the serverless implementation.

In the case of Azure Container Instances, Legion will create an Azure Container Instance in the role of Spark Worker for every executor requested.

The Spark Master is also deployed as an Azure Container Instance, and once the job has come to completion, the Spark Workers will be deleted.

This enables a true "pay per use" model for Spark.


## Usage

### Prerequisites
* Install and setup [Go](https://golang.org/doc/install)
* Install [Azure CLI 2.0](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) - For the Azure Container Instances backend

### Install the CLI ###

Download the correct Binary depending on your OS from the [Releases page](https://github.com/legion/legion/releases). <br>
Copy the binary to your bin path and you're good to go.

### Command-Line Usage

```
Run Serverless On-Demand Spark Jobs on Cloud Native platforms

Usage:
  0.0.1 [command]

Available Commands:
  deploy-master deploy a legion master
  help          Help about any command
  submit        submit a spark job

Flags:
  -h, --help   help for 0.0.1

Use "0.0.1 [command] --help" for more information about a command.
```

### Azure Container Instances

Create an Azure SPN:
```
$ az ad sp create-for-rbac --name ServicePrincipalName --password PASSWORD >> mycreds.json
```

Set the following environment variables:

* APP_ID - The Azure APP ID of the SPN (**Required**)
* PASSWORD - The Password of the SPN (**Required**)
* TENANT_ID - The Tenant_ID of the SPN (**Required**)
* AZURE_STORAGE_ACCOUNT_NAME (**Optional**) - Name of an Azure Storage Account, if processing data from Azure Blob Storage
* AZURE_STORAGE_ACCOUNT_KEY (**Optional**) - Key of the Azure Storage Account

<br>
All set!
To see the CLI help, just type ```$ legion``` to see the help menu.

### Examples

Deploy a Python Spark app with 4 executors of 4 vcpus and 8gb each, and a master with 2 vcpus and 2gb in Azure West US

```
$ legion submit --executor-cores 4 --executor-memory 8g --driver-cores 2 --driver-memory 2g --location westus --num-executors 4 --backend aci --file ./myapp.py
```

Use an existing Spark Master

```
$ legion submit --master <IP-ADDRESS OR FQDN> --backend aci --file ./myapp.py
```

Keep workers running after job is completed

```
$ legion submit --executor-cores 4 --executor-memory 8g --driver-cores 2 --driver-memory 2g --location westus --num-executors 4 --backend aci --file ./myapp.py --keep-alive true
```

Deploy Legion Master in Azure West US

```
$ legion deploy-master --driver-cores 2 --driver-memory 2g --location westus --backend aci
```

## Current limitations

* Python only (Java and Scala coming soon)
* Master is not deleted after job is done - only the workers
* Dynamic Allocation not supported
* Azure Data Lake Store not supported - only Azure Blob Storage (WASB)

## Development

* Setup and install [Go](https://golang.org/doc/install).
* Setup and install [Docker](https://docs.docker.com/install/).
* Run ``` go get -d ./...``` to get all dependencies.

### Adding backends

Legion supports a plugin enabled architecture that allows for creating new backends quickly. <br>
Take a look at the aci_backend.go implementation for Azure Container Instances to see an example of how to build a backend.

