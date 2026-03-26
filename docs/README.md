# Kafka Fundamental

## Workflow

### 1. Writh path

- How `log.Append` flows down through segment -> store + index in parallel.

![1774512499971](image/README/1774512499971.png)

### 2. Read path

- How `log.Read` binary-searches segments, then index -> store.

![1774513390998](image/README/1774513390998.png)

### 3. Segment roll

- The `IsMaxed()` check and how a new segment is created after very append.

![1774513620286](image/README/1774513620286.png)

### 4. Start up

- How `NewLog` rebuilds state from disk on restart.

![1774514240121](image/README/1774514240121.png)