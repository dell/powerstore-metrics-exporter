# README
#### About
# Metrics Exporter for Dell PowerStore

This exporter collects metrics from multiple PowerStore systems using PowerStore's RESTful API. It supports Prometheus or Zabbix for data collection and Grafana for data visualization. This exporter has been tested with PowerStore REST API versions 1.0, 2.0, 3.5; Zabbix version 6.0LTS, Prometheus version 2.39.1, and Grafana version 9.3.8.

#### build
This project is to be built using a Go environment.

```
cd PowerStoreExporter
go build -o PowerStoreExporter
```
#### Run
The exporter config file is ./config.yml and can be changed to point to another port other than the default of 9010. It is strongly recommended to create an operator user role in PowerStore, then update the storeageList section with the IP address and username/password details of the PowerStore(s).

```
./PowerStoreExporter -c config.yml
```


#### Collect
base path: http://{#Exporter IP}:{#Exporter Port}/metrics

```
Cluster              /{#PowerStoreIP}/cluster
Appliance            /{#PowerStoreIP}/appliance
Capacity             /{#PowerStoreIP}/capacity
Hardware             /{#PowerStoreIP}/hardware
Volume               /{#PowerStoreIP}/volume
VolumeGroup          /{#PowerStoreIP}/volumeGroup
Port                 /{#PowerStoreIP}/port
Nas                  /{#PowerStoreIP}/nas
FileSystem           /{#PowerStoreIP}/file
```
Sample: http://127.0.0.1:9010/metrics/10.0.0.1/Cluster

You can choose either Prometheus or Zabbix to collect/scrape metrics, then use Grafana to render/visualize the metrics.
For Prometheus the flow would be: PowerStore(s) --> exporter --> multiple targets --> Prometheus scrape jobs --> Prometheus --> Grafana
For Zabbix the flow would be: PowerStore(s) --> exporter --> multiple targets --> [ Create PowerStore host in Zabbix --> Link this host with PowerStore Zabbix template --> Scrape targets by Zabbix http client --> Zabbix DB --> Zabbix API] --> Grafana


#### Prometheus + Grafana

Add ./templates/prometheus/prometheus.yml to all jobs in your Prometheus .yml config file, then restart your Prometheus instance or reload. You can update scrape interval time to support your application monitoring requirements. We use Grafana to render metrics collected by Prometheus.

#### Zabbix and Grafana
When you create a host in Zabbix, use ./templates/zabbix/zbx_exporter_tempaltes.yaml to link PowerStore(s) to the Zabbix host. We use Grafana to render metrics collected by Zabbix. You can also create dashboards in Zabbix directly.

