To test the simple JSON WAL server, launch it: 
```
go run cmd/server/main.go
```

In other tab do POST and GET requests:
```
curl -X POST localhost:8080 -d '{"record": {"value": "BLA1234567FOOBAR"}}'
curl -X POST localhost:8080 -d '{"record": {"value": "FOO1234567FOOBAR"}}'
curl -X POST localhost:8080 -d '{"record": {"value": "XXX1234567FOOBAR"}}'

curl -X GET localhost:8080 -d '{"offset": 0}'
curl -X GET localhost:8080 -d '{"offset": 1}'
curl -X GET localhost:8080 -d '{"offset": 2}'
```