-- migrations/000006_add_password_reset_tokens.down.sql

DROP TABLE IF EXISTS auth.password_reset_tokens;
