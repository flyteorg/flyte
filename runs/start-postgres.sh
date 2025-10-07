#!/bin/bash

# Stop any existing container
docker stop flyte-runs-postgres 2>/dev/null || true
docker rm flyte-runs-postgres 2>/dev/null || true

# Start PostgreSQL with the correct password
echo "Starting PostgreSQL with Docker..."
docker run --name flyte-runs-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=flyte_runs \
  -p 5433:5432 \
  -d postgres:15

echo "Waiting for PostgreSQL to be ready..."
sleep 3

# Check if it's ready
if docker exec flyte-runs-postgres pg_isready -U postgres; then
  echo "✓ PostgreSQL is ready!"
  echo "  Host: localhost"
  echo "  Port: 5433"
  echo "  Database: flyte_runs"
  echo "  Username: postgres"
  echo "  Password: postgres"
else
  echo "❌ PostgreSQL failed to start"
  exit 1
fi
