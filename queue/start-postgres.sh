#!/bin/bash

# Stop any existing container
docker stop flyte-queue-postgres 2>/dev/null || true
docker rm flyte-queue-postgres 2>/dev/null || true

# Start PostgreSQL with the correct password
echo "Starting PostgreSQL with Docker..."
docker run --name flyte-queue-postgres \
  -e POSTGRES_PASSWORD=mysecretpassword \
  -e POSTGRES_DB=flyte_queue \
  -p 5432:5432 \
  -d postgres:15

echo "Waiting for PostgreSQL to be ready..."
sleep 3

# Check if it's ready
if docker exec flyte-queue-postgres pg_isready -U postgres; then
  echo "✓ PostgreSQL is ready!"
  echo "  Host: localhost"
  echo "  Port: 5432"
  echo "  Database: flyte_queue"
  echo "  Username: postgres"
  echo "  Password: mysecretpassword"
else
  echo "❌ PostgreSQL failed to start"
  exit 1
fi
