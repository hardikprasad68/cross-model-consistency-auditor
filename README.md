\# Cross-Model Consistency Auditor (CMCA)



A tool that measures how consistently a single prompt behaves across

different LLMs — quantifying divergence across semantic, structural, and

behavioral axes, and classifying each prompt as portable or fragile.



\*\*Core question this answers:\*\* \*"Will this prompt behave the same if you

switch the model underneath it?"\*



\## Architecture
YAML prompt suite → Go orchestrator (concurrent fan-out to N models)

↓

Python divergence engine (FastAPI + sentence-transformers)

↓

Classification (portable / cosmetically / structurally /

behaviorally fragile)

↓

DuckDB storage (runs.jsonl → cmca.duckdb)

↓

HTML report (color-coded heatmap)



\## Tech Stack



\- \*\*Go\*\* — concurrent orchestration, provider adapters (Claude, GPT, Gemini, Ollama)

\- \*\*Python + FastAPI\*\* — embedding-based divergence scoring service

\- \*\*sentence-transformers\*\* — semantic similarity via embeddings

\- \*\*DuckDB\*\* — structured storage and SQL analytics

\- \*\*Ollama\*\* — local LLM inference (Llama 3.1, Llama 3.2, Mistral, Gemma2, Qwen2.5)



\## Project Structure
cmca/

├── go-orchestrator/ # Go: provider adapters, runner, scoring client, storage

├── py-divergence/ # Python: FastAPI scoring service, DuckDB loader, report generator

├── prompts/ # YAML test-case definitions

├── data/ # Generated: runs.jsonl, cmca.duckdb

├── reports/ # Generated: findings\_report.md, report.html

├── .env.example # Template for API keys (copy to .env, fill in your own)

└── .gitignore



\## Setup



\### Prerequisites

\- Go 1.21+

\- Python 3.10+

\- \[Ollama](https://ollama.com) installed and running



\### 1. Pull local models

```bash

ollama pull llama3.1

ollama pull llama3.2

ollama pull mistral

ollama pull gemma2

ollama pull qwen2.5

```



\### 2. Set up environment variables

```bash

cp .env.example .env

\# Edit .env and add your API keys if you want to test Claude/GPT/Gemini adapters

\# (not required -- the project runs fully on local Ollama models without any keys)

```



\### 3. Set up the Python scoring service

```bash

cd py-divergence

python -m venv venv

source venv/bin/activate     # or .\\venv\\Scripts\\Activate.ps1 on Windows

pip install sentence-transformers fastapi uvicorn duckdb

uvicorn main:app --port 8000

```

Leave this running in its own terminal.



\### 4. Set up and run the Go orchestrator

In a second terminal:

```bash

cd go-orchestrator

go mod tidy

go run .

```



This loads every `.yaml` file in `../prompts`, runs each prompt concurrently

across all 5 local models, scores the divergence, classifies the result, and

appends everything to `../data/runs.jsonl`.



\### 5. Query and visualize results

```bash

cd py-divergence

python load\_and\_query.py      # SQL summary in the terminal

python generate\_report.py     # generates ../reports/report.html

```



\## How It Works



1\. \*\*Provider adapters\*\* (`go-orchestrator/\*\_provider.go`) implement a shared

&#x20;  `Provider` interface so any LLM (local or commercial) can be called the

&#x20;  same way.

2\. \*\*Concurrent runner\*\* (`runner.go`) fans a single prompt out to all

&#x20;  providers at once using goroutines.

3\. \*\*Divergence engine\*\* (`py-divergence/main.py`) scores outputs on three

&#x20;  independent axes: semantic (embedding cosine distance), structural (JSON

&#x20;  validity / word-count variance), and behavioral (refusal/hedging detection).

4\. \*\*Classification\*\* (`classify.go`) turns the three scores into one label.

5\. \*\*Storage\*\* (`storage.go`) appends every run to a JSONL file, loaded into

&#x20;  DuckDB for querying.

6\. \*\*Reporting\*\* (`generate\_report.py`) renders a static, color-coded HTML

&#x20;  heatmap.



\## Results \& Findings



See \[`reports/findings\_report.md`](reports/findings\_report.md) for the full

write-up, results table, key findings, limitations, and future work.



\## Known Limitations



\- Structural divergence falls back to word-count variance for unconstrained

&#x20; prompts (e.g. code generation) rather than validating actual correctness.

\- Behavioral divergence relies on a fixed keyword list for refusal detection.

\- Commercial API adapters (Claude, GPT, Gemini) are implemented and verified

&#x20; to compile correctly, but full live testing was limited by provider-side

&#x20; billing/capacity constraints at development time.



\## Future Work



\- Expand the test suite beyond 1 prompt per category

\- Replace the structural word-count proxy with real format validators

\- Enable full commercial-model comparison

\- Add a terminal UI for live result browsing

