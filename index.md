<h1 align="center">BackStore - Backup and Restore PVC Custom k8s Controller</h1>

---


## 📝 Table of Contents

- [About](#about)
- [Getting Started](#getting_started)
- [Running the Code](#run)
- [Authors](#authors)
- [Acknowledgments](#acknowledgement)

## 🧐 About <a name = "about"></a>

The BackStore custom kubernetes controller is written primarily in go lang. This controller explicitly keeps a watch on newly created CRs for Backupping and Restoring in all Namespaces,<br>
And as soon as a new CR is created, our controller will create Backup or Restore Data of PVC based on the CR.

## 🏁 Getting Started <a name = "getting_started"></a>

These instructions will get you the project up and running on your local machine for development and testing purposes. See [Running the Code](#run) for notes on how to deploy the project on a Local System or on a Kubernetes Server.

### Prerequisites

To run the BackStore Controller on Local System, first we need to install following Software Dependencies.

- [Go](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)

Once above Dependencies are installed we can move with [further steps](#installing)

### Installing <a name = "installing"></a>

A step by step series of examples that tell you how to get a development env running.

#### Step 1: Install Project related Dependencies
```
go mod tidy
```

#### Step 2: Running a 2 Node Mock Kubernetes Server Locally using minikube
```
minikube start --nodes 2
```

#### Step 3: Enable volumesnapshots and csi-hostpath-driver addons:
```
minikube addons enable volumesnapshots
minikube addons enable csi-hostpath-driver
kubectl create -f manifests/rbac-csi-snapshotter.yaml
kubectl create -f manifests/rbac-external-snapshotter.yaml
```

#### Step 4: Setting Up Environmental Variables

Set up the Environmental variables according to your needs. The Application will run with defaults as mentioned in the following table

| Environmental Variable  | Usage                                | Default Values    |
|-------------------------|--------------------------------------|-------------------|
| BACKSTORE_RESTORE_QUEUE | Queue for holding Restore CR objects | BACKSTORE_RESTORE |
| BACKSTORE_BACKUP_QUEUE  | Queue for holding Backup CR objects  | BACKSTORE_BACKUP  |


#### Step 5: Creating MySQL Deployments, PV and PVC
```
kubectl create -f manifests/mysql-secret.yaml
kubectl create -f manifests/mysql-storage.yaml
kubectl create -f manifests/mysql-deployment.yaml
```

#### Step 5: Creating CRDs for Backup and Restore
```
kubectl create -f manifests/backup-crd.yaml
kubectl create -f manifests/restore-crd.yaml
```


## 🔧 Running the Code <a name = "run"></a>

To Run the BackStore Controller on local machine, Open a terminal in the Project and run following command
```
go build
```
```
./backstore
```

Insert Some Data in Mysql PV,PVC

```
kubectl create -f manifests/backup.yaml
```
Delete Data from Mysql PV,PVC
```
kubectl create -f manifests/restore.yaml
```


## ✍️ Authors <a name = "authors"></a>

- [@r4rajat](https://github.com/r4rajat) - Implementation

## 🎉 Acknowledgements <a name = "acknowledgement"></a>

- References
    - https://pkg.go.dev/k8s.io/client-go
    - https://pkg.go.dev/k8s.io/apimachinery
    - https://pkg.go.dev/github.com/mitchellh/go-homedir
    - https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/
    - https://pkg.go.dev/github.com/kubernetes-csi/external-snapshotter/v6
