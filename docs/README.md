
## Build 

```bash
# NOLINT=true make               # build binary: bin/puller

Or

# NOLINT=true make image         # build docker image: bbklab/puller:latest
```

## Run
```bash
# ./bin/puller serve --listen :9006 --docker-socket /var/run/docker.sock

Or

# docker run -d --name=puller \
	-e LISTEN_ADDR=:9006 \
	-e DOCKER_SOCKET=/var/run/docker.sock \
	--net=host \
	-v=/var/run/docker.sock:/var/run/docker.sock \
	bbklab/puller
```

## API

### Ping
`GET /ping`  -  ping

```liquid
OK
```

### Debug
`GET /debug/dump`  -  dump debug informations

Response:
```json
{
  "configs": {
    "listen": ":9006",
    "docker_socket": "/var/run/docker.sock"
  },
  "num_fds": 9,
  "start_at": "2018-10-18T11:24:34+08:00"
}
```

### Pull
`POST /pull`  -  request node docker to pull images

Request:
```json
{
  "concurrency": 10,
  "retry": 3,
  "images": [
    {
      "image": "busybox",
      "tag": "latest",
      "project": "p1"
    },
    {
      "image": "tom/private-image",
      "tag": "latest",
      "project": "p2",
      "auth_config": {
        "username": "tom",
        "password": "tom_password",
      }
    },
    {
      "image": "bbklab/abc-master",
      "tag": "latest",
      "project": "p3",
      "auth_config": {
        "identitytoken": "xxxxxxx"
      }
    }
  ]
}
```

Response:
```json
- 200 全部成功
- 500 至少一个镜像拉取失败

{
  "success": [
    {
      "image": "tom/private-image",
      "tag": "latest",
      "project": "p2",
      "cost": "3.539696838s",
      "retried": 1,
      "errmsg": ""
    },
    {
      "image": "busybox",
      "tag": "latest",
      "project": "p1",
      "cost": "5.112341751s",
      "retried": 1,
      "errmsg": ""
    }
  ],
  "failure": [
    {
      "image": "bbklab/abc-master",
      "tag": "latest",
      "project": "p2",
      "cost": "7.884867327s",
      "retried": 3,
      "errmsg": "Error response from daemon: Get https://registry-1.docker.io/v2/bbklab/abc-master/manifests/latest: unauthorized: incorrect username or password"
    }
  ],
  "startAt": "2018-10-18T11:10:24.36290826+08:00",
  "cost": "5.112453988s"
}
```
