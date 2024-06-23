
## Docker
docker build -t my-envoy-with-lua .
docker run -d -p 10000:10000 -p 9901:9901 --name envoy-proxy-with-lua my-envoy-with-lua
docker logs envoy-proxy-with-lua
curl -v http://localhost:10000/

## Docker Compose 
> docker-compose up --build

## Output 
```shell
curl -v http://localhost:10000/
*   Trying [::1]:10000...
* Connected to localhost (::1) port 10000
> GET / HTTP/1.1
> Host: localhost:10000
> User-Agent: curl/8.4.0
> Accept: */*
> 
< HTTP/1.1 200 OK
< server: envoy
< date: Sun, 23 Jun 2024 15:46:00 GMT
< content-type: text/plain
< x-envoy-upstream-service-time: 7
< x-response-header: response_value
< transfer-encoding: chunked
< 
* Connection #0 to host localhost left intact
Hello, World!%           
```