# specialist-talk-api

###
```bash
docker run -d -p 9000:9000 -p 9001:9001 --restart always --name minio -e MINIO_ROOT_USER=admin -e MINIO_ROOT_PASSWORD=I756mab9yEmK quay.io/minio/minio server /data --console-address ":9001"
```