# syntax=docker/dockerfile:1
FROM postgres:latest
COPY init.sql /docker-entrypoint-initdb.d/