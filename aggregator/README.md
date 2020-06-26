# aggregator

## running

### compile

```
bazel build //aggregator
```

### init jwt token

```
ssh-keygen -t rsa -b 4096 -m PEM -f jwtRS256.key # Don't add passphrase
openssl rsa -in jwtRS256.key -pubout -outform PEM -out jwtRS256.key.pub
```

## testing

### login

```
grpcurl -d '{"email":"i@adamyi.com","password":"123"}' -plaintext 127.0.0.1:4200 productimon.svc.DataAggregator.Login
```

### extend token

```
grpcurl -H 'Authorization: {{token}}' -plaintext 127.0.0.1:4200 productimon.svc.DataAggregator.ExtendToken
```

### check logged in user details

```
grpcurl -H 'Authorization: {{token}}' -plaintext 127.0.0.1:4200 productimon.svc.DataAggregator.UserDetails
```
