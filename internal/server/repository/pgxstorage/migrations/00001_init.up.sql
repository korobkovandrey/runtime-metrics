CREATE TABLE IF NOT EXISTS metrics (
   type VARCHAR(255) NOT NULL,
   id VARCHAR(255) NOT NULL,
   value DOUBLE PRECISION NULL,
   delta BIGINT NULL,
   PRIMARY KEY (type, id)
);
