# PostgreSQL Dockerfile for E-commerce Application
FROM postgres:16-alpine

# Set environment variables for PostgreSQL
ENV POSTGRES_DB=mandi
ENV POSTGRES_USER=rohitjangid
ENV POSTGRES_PASSWORD=12345

# Copy initialization scripts
COPY initdb/ /docker-entrypoint-initdb.d/

# Expose PostgreSQL port
EXPOSE 5432

# The postgres image automatically runs initdb scripts on first startup
# These scripts will create tables, indexes, and initial data