#!/bin/bash

# Health Check
echo "==== Health Check ===="
curl -s http://localhost:8080/health | jq

# Get All Users
echo "\n==== Get All Users ===="
curl -s http://localhost:8080/api/users | jq

# Get User by ID
echo "\n==== Get User by ID (1) ===="
curl -s http://localhost:8080/api/users/1 | jq

# Get User with Profile (calls user-profile-service - generates service dependencies)
echo "\n==== Get User with Profile (1) - Cross Service Call ===="
curl -s http://localhost:8080/api/users-with-profile/1 | jq

# Get All Products
echo "\n==== Get All Products ===="
curl -s http://localhost:8080/api/products | jq

# Get Product by ID
echo "\n==== Get Product by ID (1) ===="
curl -s http://localhost:8080/api/products/1 | jq

# Create Order
echo "\n==== Create Order ===="
curl -s -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "product_id": 1, "quantity": 2}' | jq

echo "\nDemo requests completed!"
