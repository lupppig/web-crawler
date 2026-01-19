# 🕷️ Go Web Crawler

This is a concurrent web crawler built using **Go**, a custom HTML parser, and **MongoDB** for data storage. It utilizes a **worker pool** pattern to efficiently crawl pages, extracts useful information, and stores it in a structured format.

---

## 🧭 Design Decisions

### Breadth-First Search (BFS)
I chose **BFS** to prioritize discovering top-level pages before diving deeper, which better simulates typical user exploration and ensures broader coverage of the site structure.

### Worker Pool
Instead of spawning a goroutine for every URL (which can be uncontrolled), I implemented a **worker pool**. This allows limiting the number of concurrent connections, respecting system resources and the target server's limits.

### Custom HTML Parser
The project uses `golang.org/x/net/html` for tokenization, but the parsing logic is custom-built to efficiently extract titles, body text, and links while skipping irrelevant tags (scripts, styles, etc.).

---

## 📌 Features

*   **Concurrency**: Uses a worker pool for parallel processing.
*   **Politeness**: Checks `robots.txt` before crawling.
*   **Data Persistence**: Stores crawled metadata and content in MongoDB.
*   **Metrics**: Tracks and logs crawl statistics (duration, pages/sec, success/failure rates) upon completion.
*   **Dockerized**: Fully containerized with Docker and Docker Compose for easy setup.

---

## 🛠️ Project Structure

The project follows a standard Go project layout:

*   `cmd/crawler`: Application entry point.
*   `internal/crawler`: Core crawler logic and worker pool implementation.
*   `internal/queue`: Thread-safe URL queue.
*   `internal/storage`: MongoDB storage implementation.
*   `internal/parser`: HTML parsing logic.

---

## 🚀 How to Run

### Prerequisites
*   Docker & Docker Compose (recommended)
*   **OR** Go 1.23+ and a running MongoDB instance

### Using Docker (Recommended)

The easiest way to run the crawler is using the provided `Makefile` and Docker Compose:

1.  **Start the Crawler:**
    ```bash
    make docker-up
    ```
    This will build the image, start MongoDB, and begin crawling the seed URL.

2.  **Stop the Crawler:**
    ```bash
    make docker-down
    ```

### Running Locally

1.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

2.  **Set up Environment:**
    Copy the example environment file:
    ```bash
    cp .env.example .env
    ```
    Update `.env` with your MongoDB credentials if necessary.

3.  **Run the application:**
    ```bash
    make run
    # OR
    go run cmd/crawler/main.go
    ```

---

## � Metrics

When the crawl finishes (or is interrupted), the crawler logs statistics to the console:

```text
--- Crawl Statistics ---
Total Duration: 2m15s
Total Pages Crawled: 342
Successful Requests: 340
Failed Requests: 2
Average Time Per Page: 0.3947 seconds
```

---

## 🧪 Testing

Run the test suite using:

```bash
make test
```

---

## 📝 Tech Stack

*   **Language**: Go 1.23
*   **Database**: MongoDB
*   **Containerization**: Docker
*   **Orchestration**: Docker Compose
