# calredact
Pulls calendar from CalDAV backend, redacts data, serves as `.ics`.

```
BACKEND=https://caldav
USER=sus
```

Password from `secrets/password` (Container: `/secrets/password`)

```
http://localhost:8080/redacted.ics?eventName=Generic Title
```
