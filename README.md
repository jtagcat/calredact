# calredact
Pulls calendar from CalDAV backend, redacts data, serves as `.ics`.

```
BACKEND=https://caldav
USER=sus
IGNORE=event1¤event two¤Third thing
```

- Password is read from `secrets/password` (Container: `/secrets/password`)
- Auth key is read from `secrets/authkey` (Container: `/secrets/authkey`)

```
http://localhost:8080/redacted.ics?auth=authKeyHere&eventName=Generic Title
```

## Updating dependencies
```
go get -u ./...
go mod vendor
```
(+ commit + docker build)
