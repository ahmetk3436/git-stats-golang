# Prometheus Installation

In this scenario, we will start with the installation of Prometheus. After that, we will proceed with the installation of Grafana and make the necessary connections to see a real-world example.

## Docker
Since we will be working with Docker, it is required to have Docker installed on our machine. Follow the instructions below to install Docker if it is not already present.

```bash
docker --version
```
If the above command does not show the Docker version, follow the steps below.

Update the system and add Docker.

```bash
apk update && apk add docker
```

Start the Docker service on the system and add the necessary function to restart it when the system reboots.
```bash
service docker start
rc-update add docker boot
```

In some cases, adding Docker to the user may be necessary. If we do not see the output in the next step, run the following command.

```bash
addgroup $USER docker
```
If all our commands are successful, we should see the "Hello, World!" output.

```bash
docker run hello-world
```
Pull the Prometheus image.
```bash
docker pull prom/prometheus
```
Start the Prometheus container.

```bash
docker run -d -p 9090:9090 --name prometheus prom/prometheus
```
And finally, our Prometheus is up and running!

By following these steps, you can successfully run Prometheus on Docker.