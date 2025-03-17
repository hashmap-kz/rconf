# Usage example (pkey-based auth)

### Run docker images

```bash
docker compose up --build -d
```

### Ensure pkey mod

```bash
chmod 0600 id_ed25519
```

### Run rconf

```bash
rconf \
  -i id_ed25519 \
  -f ./scripts/ \
  -H 'root@localhost:12222' \
  -H 'root@localhost:12223?sudo=true'
```
