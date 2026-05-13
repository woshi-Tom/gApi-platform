# gAPI Platform v{VERSION}

## Quick Start

1. **Configure**: Copy `.env.example` to `.env` and edit passwords
2. **Start**: Run `./start.sh` (Linux/Mac) or `start.bat` (Windows)
3. **Access**:
   - Frontend: http://localhost:5173
   - Admin: http://localhost:5174
   - API: http://localhost:8080
   - API Docs: http://localhost:8080/swagger/index.html

## Requirements

- [Docker](https://docs.docker.com/get-docker/) (with Compose)

## Configuration

Edit `.env` to set:

| Variable | Description | Required |
|----------|-------------|----------|
| POSTGRES_PASSWORD | PostgreSQL password | Yes |
| REDIS_PASSWORD | Redis password | Yes |
| RABBITMQ_PASSWORD | RabbitMQ password | Yes |
| JWT_SECRET | JWT key (min 32 chars) | Yes |
| ENCRYPT_KEY | Encryption key (min 32 chars) | Yes |

## Directory Structure

Linux package (`*-linux-amd64.tar.gz`):
```
gapi-platform-vX.X.X-linux/
├── bin/linux/amd64/gapi-server    # Linux binary
├── frontend/                       # User frontend (Nginx)
├── admin/                          # Admin panel (Nginx)
├── docker-compose.yml              # Docker Compose config
├── Dockerfile.backend              # Backend Docker image
├── nginx.conf                      # Nginx config
├── .env.example                    # Environment template
├── config.yaml.example             # Backend config template
├── start.sh                        # Startup script (Linux/Mac)
└── README.md                       # This file
```

Windows package (`*-windows-amd64.zip`):
```
gapi-platform-vX.X.X-windows/
├── bin/windows/amd64/gapi-server.exe  # Windows binary
├── frontend/                           # User frontend (Nginx)
├── admin/                              # Admin panel (Nginx)
├── docker-compose.yml                  # Docker Compose config
├── Dockerfile.backend                  # Backend Docker image
├── nginx.conf                          # Nginx config
├── .env.example                        # Environment template
├── config.yaml.example                 # Backend config template
├── start.bat                           # Startup script (Windows)
└── README.md                           # This file
```
