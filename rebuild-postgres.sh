#!/bin/bash
echo "Rebuilding PostgreSQL container with fresh init.sql..."
docker-compose down -v
docker-compose build --no-cache postgres
docker-compose up -d postgres
echo "Done! PostgreSQL is ready with updated init.sql"