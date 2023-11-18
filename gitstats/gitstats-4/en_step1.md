# Git Stats Demo

## Objective
In this demonstration, we will explore statistics on Git. We will identify the Lines Of Code (LOC) in a repository, finding all code lines. Additionally, we will observe the number of additions and deletions made by each developer.

```bash
git clone https://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@github.com/ahmetk3436/git-stats-golang
```

## Navigate to the Directory
```bash
cd git-stats-golang
```
## Run Docker Compose
```bash
docker compose up -d
```

## Make Changes

If you make any changes, you can rebuild by running the following command:

```bash
docker compose up --build -d
```
This Docker Compose file will deploy a backend, Prometheus, and Grafana server. These systems are automatically interconnected. The /metrics endpoint in the backend contains Prometheus metric outputs. Additionally, under the yaml folder, a dashboard tracking Prometheus, Grafana, and Golang metrics will be available.

## Explore the Data

Feel free to navigate to port 80 in your current environment to explore the data of my repositories. Enjoy!