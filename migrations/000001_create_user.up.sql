CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  mail VARCHAR(128) NOT NULL,
  password VARCHAR(60) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP,
  UNIQUE(mail)
);