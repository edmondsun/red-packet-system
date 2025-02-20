# 🧧Red Packet System (Built with Golang, MySQL, Redis, Kafka)

## Overview
This is a high-performance red packet system built with Golang, MySQL, Redis Cluster, and Kafka. It demonstrates:
- High-concurrency red packet grabbing.
- Distributed transaction logging via Kafka (Producer/Consumer).
- Master-Slave Replication in MySQL, using dbresolver for read/write splitting.
- Redis caching, Bloom Filter for anti-penetration, RedLock for distributed locking, and Lua scripts for atomic stock decrement.
- Robust architecture to handle million-level concurrency, leveraging Docker + Docker Compose for container orchestration.

---
## 🏗 Project Structure
```
red-packet-system/
├── cmd/                    # Entry points for different services
│   ├── kafka/
│   │   └── worker.go        # Kafka consumer worker
│   ├── server/
│   │   └── server.go        # API server main entry point
│
├── config/                  # Configuration files
│   ├── config.go            # Loads environment variables and system configurations (singleton)
│
├── db/                      # Database-related logic
│   ├── migrations/          # SQL migration scripts
│   │   ├── 000001_red_packets.up.sql      # Red packet table
│   │   ├── 000002_red_packet_logs.up.sql  # Red packet transaction logs
│   │   ├── 000003_users.up.sql            # Users table
│   ├── mysql.go             # GORM + dbresolver for read/write splitting
│   ├── seed.go              # Database seed data
│
├── kafka/                   # Kafka producer and consumer
│   ├── consumer.go          # Kafka consumer logic
│   ├── producer.go          # Kafka producer logic
│   ├── utils.go             # Helper functions for retries and error handling
│
├── model/                   # Data models (GORM-based)
│   ├── red_packet.go        # RedPacket struct and ORM mappings
│   ├── red_packet_log.go    # RedPacketLog struct for transaction logs
│   ├── user.go              # User struct
│
├── pkg/                     # Utility libraries
│   ├── logger/
│   │   ├── logger.go        # Logger singleton for structured logging  (singleton)
│
├── redisclient/             # Redis cluster and Redlock-based distributed locks
│   ├── redis.go             # Redis connection and operations
│
├── routes/                  # API routes and endpoint definitions
│   ├── router.go            # Gin router setup
│
├── service/                 # Business logic and services
│   ├── red_packet_service.go # Core logic for grabbing red packets
│
├── api/                     # API handlers
│   ├── handler.go           # HTTP handlers for API endpoints
│
├── nginx/                   # Nginx configuration
│   ├── nginx.conf           # Load balancing and reverse proxy settings
│
├── .env                     # Environment variables
├── Dockerfile               # Docker setup for Multi-stage build (Go -> minimal runtime)
├── docker-compose.yml       # Docker Compose setup for microservices (MySQL, Redis Cluster, Kafka, etc.)
├── go.mod                   # Go module dependencies
├── go.sum                   # Go module checksums
├── README.md                # Project documentation
```

---
## 📈 Million-Level Concurrency Architecture
```
+--------------------------------+
|  Users (APP / Frontend)        |
+--------------------------------+
            ↓
+-----------------------------------------------+
|  Nginx (Load Balancing / Rate Limit -- TODO)  |
+-----------------------------------------------+
            ↓
+--------------------------------+
|  API Server (Golang + Gin)     |
+--------------------------------+
            ↓
+-------------------------------------------------------+
| Redis Cluster (Lua Scripts & Bloom Filter & RedLock)  |
|  Kafka (Multi-Partition Consumers)                    |
|  MySQL (Master-Slave Read/Write Splitting)            |
+-------------------------------------------------------+
```

---
## 🔑 Key Techniques & Features

### **1. Redis Cluster**
- Lua Scripts for atomic stock decrement (prevents race conditions).
- RedLock (distributed lock) ensures a single user can’t over-grab the same red packet concurrently.
- Bloom Filter to avoid cache penetration. If an ID isn’t in the Bloom Filter, we skip querying MySQL.
- Randomized TTL to mitigate cache avalanche (spreading expiration times so keys don’t all expire simultaneously).

### **2. MySQL Master-Slave Replication**
- Using GORM’s dbresolver plugin for read/write splitting: writes go to the master; reads go to the slave to scale read traffic.
- Seed script to populate test data.
- Transactional updates for red packet counts, ensuring strong consistency in MySQL itself.

### **3. Kafka**
- Producer: sends transaction events (user + amount) to Kafka.
- Consumer: runs in a separate worker to update user balances asynchronously.
- retryWithBackoff logic ensures robust error handling and prevents repeated consumption.
- Leverages partitioning to distribute load among consumers in a group.

### **4. Singleton Patterns**
- Config (using sync.Once to load .env or environment variables).
- Logger (shared logger instance).
- DB connection (GORM).
- Redis client.
- Minimizes overhead and ensures consistent usage across the codebase.

### **5. Graceful Shutdown**
- Listens for signals like SIGTERM, gracefully stops the HTTP server, flushes logs, closes DB connections, and stops Kafka consumption.

### **6. Docker & Docker Compose**
- Multi-stage Go build: minimal final image with only the compiled binaries.
- docker-compose.yml orchestrates MySQL (master + slave), Redis cluster, Kafka + Zookeeper, and the application containers.
- Health checks for MySQL, Kafka, and the Go services.

---
## 🚀 Quick Start

### **1. Prerequisites**
Ensure you have the following installed on your machine:
- **Docker** & **Docker Compose**
- **Golang 1.22+**
- **MySQL 5.7**
- **Redis**
- **Kafka & Zookeeper**

### **2. Setup & Installation**
```
git clone https://github.com/your-repo/red-packet-system.git
cd red-packet-system
docker-compose up -d --build
```

### **3. Database Migration & fake data**
Run migrations:
```
docker exec -it server-api /app/migrate -path=db/migrations \
  -database "mysql://root:123456@tcp(mysql-master:3306)/red_packet_db" up
```

Seed data (inserts test users & red packets):
```
docker exec -it server-api /app/server-api --seed
```

### **4. Configure MySQL Slave**
```
docker-compose -f docker-compose.yml exec mysql-slave bash

# Inside container:
mysql -uroot -p123456

# Set up replication
CHANGE MASTER TO MASTER_HOST='mysql-master', MASTER_USER='root', MASTER_PASSWORD='123456', MASTER_LOG_FILE='mysql-bin.000001', MASTER_LOG_POS=4;

# Ativate
START SLAVE;

# Check States
SHOW SLAVE STATUS\G;
```

### **5. Redis Cluster Setup**
```
# Redis Setting
docker-compose -f docker-compose.yml exec redis-cluster-1 bash

# Inside container:
redis-cli --cluster create \
redis-cluster-1:6379 \
redis-cluster-2:6379 \
redis-cluster-3:6379 \
--cluster-replicas 0

# Check cluster info
redis-cli cluster info
redis-cli cluster nodes
```

### **6. kfaka Usage**
Consumer (manual check):
```
docker exec -it kafka bash -c "/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic red_packet_transactions --from-beginning"
```

Producer (for test messages):
```
docker exec -it kafka bash -c " /opt/kafka/bin/kafka-console-producer.sh --bootstrap-server localhost:9092 --topic red_packet_transactions"

# Then type a test message:
> 1,1,43.75
```

### **7. API Endpoints**
Health Check:
```
curl -X GET "http://localhost:8080"

{"message":"Red Packet System is running!"}
```

Grab a Red Packet:
```
curl -X GET "http://localhost:8080/grab?user_id=1&red_packet_id=1"

{
  "message": "Red packet grabbed successfully",
  "amount": 5.67
}
```

### **8. Logs & Monitoring**
```
# API logs
docker logs -f server-api

# Kafka consumer logs
docker logs -f kafka-worker
```