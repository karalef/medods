# MEDODS Test Task

## Prerequisites

Docker and Docker Compose

## Run

Copy the .env.example as .env
```sh
cp .env.example .env
```

Generate a random base64 secret (16 or more bytes for security reasons)
```sh
openssl rand -base64 32
```

Paste the generated secret to the variable in .env file
```sh
SECRET=your_generated_secret
```

Start with docker compose
```sh
docker compose --env-file ./.env up -d
```
