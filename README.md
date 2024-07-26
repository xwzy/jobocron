# jobocron: Distributed Task Scheduling System

jobocron is a powerful, flexible, and scalable distributed task scheduling system written in Go. It's designed to handle complex job scheduling scenarios in modern distributed environments.

## Features

- **Web-based Management**: Easy CRUD operations for jobs through a user-friendly web interface.
- **Dynamic Job Control**: Start, stop, or terminate running jobs in real-time.
- **High Availability**: Both the scheduling center and executors support cluster deployment for HA.
- **Flexible Trigger Strategies**: Support for Cron, fixed interval, fixed delay, API, manual, and parent-child job triggering.
- **Smart Routing**: Various job routing strategies including round-robin, random, consistent hashing, least frequently used, failover, etc.
- **Fault Tolerance**: Automatic failover and configurable retry mechanisms for failed jobs.
- **Real-time Monitoring**: Track job progress and view logs in real-time.
- **Distributed Execution**: Support for job sharding and dynamic scaling of executor clusters.
- **Cross-language Support**: RESTful APIs allow integration with various programming languages.
- **Security**: Built-in support for data encryption and fine-grained access control.

## Architecture

jobocron consists of three main modules:

1. **Controller**: 
   - Scheduler: Decides which jobs to schedule.
   - Dispatcher: Determines how to execute jobs.

2. **Worker**: 
   - JobExecutor: Responsible for actual job execution.

3. **Common**: 
   - Shared components like data models, utilities, and persistence layer.

## Getting Started

### Prerequisites

- Go 1.16+
- MySQL 5.7+
- Redis 5.0+
- Kafka 2.8+

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/jobocron.git
   ```

2. Navigate to the project directory:
   ```
   cd jobocron
   ```

3. Install dependencies:
   ```
   go mod tidy
   ```

4. Build the project:
   ```
   go build -o jobocron
   ```

5. Configure the application:
   - Edit `config.yaml` to set up database, Redis, and Kafka connections.

6. Run the application:
   ```
   ./jobocron
   ```

## Usage

After starting the application, you can access the web interface at `http://localhost:8080`. 

Here's a simple example of how to create a job via the API:

```bash
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Hello World Job",
    "type": "shell",
    "script": "echo Hello, World!",
    "schedule": "0 */5 * * * ?",
    "timeout": 60
  }'
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for more details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
