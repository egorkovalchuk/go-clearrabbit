# RabbitMQ Queue Cleanup Utility
This project is a utility for cleaning all queues on RabbitMQ servers. The utility uses the RabbitMQ HTTP API to retrieve a list of queues and clear them.

## Description
The utility performs the following tasks:
1. Reading configuration from a JSON file.
2. Connecting to RabbitMQ servers.
3. Retrieving a list of queues via the RabbitMQ HTTP API.
4. Clearing all queues on the specified servers.

## Usage
Command-line Flags
* **-d** : Enable debug mode (outputs additional information).
* **-v** : Display the utility version.
* **-c** : Path to the configuration file (default: config.json).
* **-l** : Login for connecting to RabbitMQ (mandatory parameter).
* **-p** : Password for connecting to RabbitMQ (mandatory parameter).
* **-h** : Display help.

## Example of Running:
```bash
go run main.go -l login -p password -c config.json
```

## Configuration
The utility's configuration is defined in the config.json file. Below is an example of the configuration structure:

```json
{
  "ServerList": [
    "rabbitmq1:15672",
    "rabbitmq2:15672"
  ]
}
```

## Explanation:
ServerList : A list of RabbitMQ server addresses with their respective HTTP API ports.

## Logs
The utility creates a log file named report.log, which records information about the queue cleanup process.







