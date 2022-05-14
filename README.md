# all about cats

## cats app

Simple Go HTTP Server implementation for saving and fetching cats [here](server.go).
Redis DB used for saving information.

## run services

---

### on local machine:

#### start app on local:

> P.S. Run redis on docker(no need for install):

```shell
export REDIS_PASSWORD=password123
docker run -d --name redis -p 6379:6379 -e REDIS_PASSWORD=password123 bitnami/redis:latest
```

```shell
go run server.go
```

#### start app on docker:(NOT RECOMMENDED)

> P.S. Run redis on docker(no need for install):

```shell
export REDIS_PASSWORD=password123
docker run -d --name redis -p 6379:6379 -e REDIS_PASSWORD=password123 bitnami/redis:latest
```

then,

```shell
# build image by name/tag `cats-on-docker`
docker build -t cats-on-docker .

# optional check (before and after)
docker images

# run the image as a container
# map TCP port 8080 in the container to port 8082 on the Docker host.
# also, pass the REDIS_PASSWORD
docker run -d -p 8080:8080 --env REDIS_PASSWORD=password123 cats-on-docker

```

---

#### start app using docker-compose:

```shell

docker-compose up --build -d

docker-compose down

```

---

### on K8s: <TODO>

---

## access app:

### save a cat:

```shell

curl -X POST "http://localhost:8080/cats" --data '{"name": "angie", "age": 2, "housename": 12}'
curl -X POST "http://localhost:8080/cats" --data '{"name": "dylan", "age": 3, "housename": 21}'
curl -X POST "http://localhost:8080/cats" --data '{"name": "owen", "age": 3, "housename": 34}'

```

### get a cat:

```shell
curl -X GET "http://localhost:8080/cats/<id>"
```

---
