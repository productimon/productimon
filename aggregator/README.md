# aggregator

## running

### compile

This bundles the frontend as well.

**prod build** (minified js)

```
bazel build //aggregator
```

**dev build** (unminified js with sourcemap, ~10MB)

```
bazel build -c dbg //aggregator
```

### init db

```
sqlite3 db.sqlite3 < schema.sql
```

### init jwt token

```
ssh-keygen -t rsa -b 4096 -m PEM -f jwtRS256.key # Don't add passphrase
openssl rsa -in jwtRS256.key -pubout -outform PEM -out jwtRS256.key.pub
```

### run

```
bazel-bin/aggregator/aggregator_/aggregator --help
```

Visit `http://127.0.0.1:4201/`

## testing

### login

```
grpcurl -d '{"email":"test@productimon.com","password":"test"}' -plaintext 127.0.0.1:4200 productimon.svc.DataAggregator.Login
```

### extend token

```
grpcurl -H 'Authorization: {{token}}' -plaintext 127.0.0.1:4200 productimon.svc.DataAggregator.ExtendToken
```

### check logged in user details

```
grpcurl -H 'Authorization: {{token}}' -plaintext 127.0.0.1:4200 productimon.svc.DataAggregator.UserDetails
```
