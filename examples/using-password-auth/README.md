# Usage example (password-based auth)

### Run docker images

```bash
docker compose up --build -d
```

### Run rconf

```bash
rconf \
  -f ./scripts/ \
  -H 'root:root@localhost:12222' \
  -H 'root:root@localhost:12223?sudo=true'
```
