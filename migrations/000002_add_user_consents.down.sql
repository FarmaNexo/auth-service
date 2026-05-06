-- migrations/000002_add_user_consents.down.sql

DROP TABLE IF EXISTS auth.user_consents CASCADE;
