# Install hey and func-e 

```shell
brew install hey
curl https://func-e.io/install.sh | bash -s -- -b /usr/local/bin\nfunc-e run -c /path/to/envoy.yaml\n
```

# Start python server

```shell
python3 server.py
func-e run -c envoy.yaml  
hey http://localhost:10000
```





