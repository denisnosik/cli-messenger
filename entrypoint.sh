#!/bin/sh
echo "Waiting for database..."
until ./goose -dir ./sql/schema postgres "$DB_URL" up; do
    echo "Database not ready, retrying in 2s..."
    sleep 2
done

echo "Starting server..."
exec ./server