import duckdb

DB_PATH = "../data/cmca.duckdb"
JSONL_PATH = "../data/runs.jsonl"

con = duckdb.connect(DB_PATH)

con.execute(f"""
    CREATE OR REPLACE TABLE runs AS
    SELECT * FROM read_json_auto('{JSONL_PATH}')
""")

print("Loaded rows:", con.execute("SELECT COUNT(*) FROM runs").fetchone()[0])

print("\n--- Average scores by category ---")
result = con.execute("""
    SELECT category,
           ROUND(AVG(semantic_score), 3) AS avg_semantic,
           ROUND(AVG(structural_score), 3) AS avg_structural,
           ROUND(AVG(behavioral_score), 3) AS avg_behavioral,
           COUNT(DISTINCT run_id) AS num_runs
    FROM runs
    GROUP BY category
""").fetchall()
for row in result:
    print(row)

print("\n--- Classification breakdown ---")
result2 = con.execute("""
    SELECT classification, COUNT(*) AS count
    FROM runs
    GROUP BY classification
    ORDER BY count DESC
""").fetchall()
for row in result2:
    print(row)

con.close()