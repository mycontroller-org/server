# MyController

MyController.org version 2.x is in early development stage.

**WARNING:** Huge change will be expected on each commit. At this moment this version is not ready for the production environment.

### Demo
* https://demo-v2.mycontroller.org (no data available at this moment)
* username: `admin`
* password: `admin`

### configuration example files
* [resources/](resources/)

### To run
```bash
# pull the image
docker pull quay.io/mycontroller-org/mycontroller:2.0-master

# run with default configuration
docker run  -d --name mycontroller \
    -p 8080:8080 \
    quay.io/mycontroller-org/mycontroller:2.0-master

# run with advanced options with custom data mount point
docker run  --rm --name mycontroller \
    -p 8080:8080 \
    -v $PWD/mc:/mc_home \
    quay.io/mycontroller-org/mycontroller:2.0-master

# run with advanced options with custom data mount point and custom configuration options
docker run  --rm --name mycontroller \
    -p 8080:8080 \
    -v $PWD/mc:/mc_home \
    -v $PWD/mc/mycontroller.yaml:/app/mycontroller.yaml \
    quay.io/mycontroller-org/mycontroller:2.0-master
```
