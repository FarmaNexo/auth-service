-- migrations/000001_init_schema.down.sql
-- Rollback de la migración inicial

-- Eliminar tablas (en orden inverso por foreign keys)
DROP TABLE IF EXISTS auth.refresh_tokens;
DROP TABLE IF EXISTS auth.users;

-- Eliminar esquema
DROP SCHEMA IF EXISTS auth;