# Collection of usefull SQL-Statements

This statements were used with a postgreSQL database.

### Create Table

```
CREATE TABLE openldap (
  constants TEXT,
  expression TEXT,
  type TEXT,
  file TEXT,
  hash TEXT,
  id SERIAL PRIMARY KEY
);
```

```
CREATE TABLE libxml2_rules (
  "id" SERIAL PRIMARY KEY,
  "base" TEXT NOT NULL,
  "add" TEXT NOT NULL,
  "confidence" FLOAT NOT NULL,
  "support" FLOAT NOT NULL
);
```