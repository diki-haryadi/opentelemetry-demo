## Load Testing

Berikut perintah curl untuk melakukan load testing pada service:

### Mendapatkan profil pengguna
```bash
# Single request
curl -X GET http://localhost:8081/api/user-profiles/1

# Load test (100 requests)
for i in {1..100}; do
  curl -X GET http://localhost:8081/api/user-profiles/$((RANDOM % 3 + 1)) > /dev/null &
done

# Menggunakan Apache Bench (ab)
ab -n 100 -c 10 http://localhost:8081/api/user-profiles/1
```

### Health check
```bash
curl -X GET http://localhost:8081/health
```
