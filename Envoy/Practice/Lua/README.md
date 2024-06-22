docker build -t my-envoy-with-lua .


docker run -d -p 10000:10000 -p 9901:9901 --name envoy-proxy-with-lua my-envoy-with-lua

docker logs envoy-proxy-with-lua


curl -v http://localhost:10000/

