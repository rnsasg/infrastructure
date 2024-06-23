
docker build -t my-envoy .
docker run -d -p 10000:10000 -p 9901:9901 --name envoy-proxy my-envoy
 docker logs envoy-proxy

 