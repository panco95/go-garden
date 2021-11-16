docker run -it --name gateway -p 8080:8080 -p 9000:9000 gateway

docker container prune

docker image rm gateway

docker start gateway

FROM alpine
COPY gateway /app/gateway
CMD [/app/gateway","-configs=/app/configs","-runtime=/app/runtime"]

docker run -it --name gateway -v /data/containers/configs:/app/configs -v /data/containers/runtime:/app/runtime -p 8080:8080 -p 9000:9000 gateway

docker build -t gateway .