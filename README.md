# Мое решение ТЗ в компанию Messagio
описание - https://docs.google.com/document/d/13JHrzO9HuExWe_X0WrJzD8TPk3UHyTowUHCJHA5RFuY/edit?usp=sharing
## Setup
- запуск postgres и kafka через докер ,по конфигу нужно config/local.yaml;
- `go mod tidy`
- `go build cmd/migrator/main.go ./main`
- `go build cmd/msg_handler/main.go ./main`
- `go build cmd/msg_stats/main.go ./main`
- `go build cmd/msg_handler/main.go ./main`
## API

https://interstellar-spaceship-71964.postman.co/workspace/message-tk~a1ebce3e-b4d4-4d56-9ee8-1b59f7c63fce/collection/33505224-95939422-2ee1-4eb8-adde-5c79d6f222c7?action=share&creator=33505224

здесь 3 микросервиса
orchestrator - api-gatewey
msg_stats - возвращает статистику по заданному периоду
msg_handler - "обрабатывает" сообщения через Apache Kafka
