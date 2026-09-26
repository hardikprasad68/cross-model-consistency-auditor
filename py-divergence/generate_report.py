import duckdb
import html as html_lib

DB_PATH = "../data/cmca.duckdb"
JSONL_PATH = "../data/runs.jsonl"
OUTPUT_PATH = "../reports/report.html"


def score_to_color(score: float) -> str:
    score = max(0.0, min(1.0, score))
    if score < 0.5:
        r = int(255 * (score / 0.5))
        g = 200
    else:
        r = 255
        g = int(200 * (1 - (score - 0.5) / 0.5))
    return f"rgb({r},{g},80)"


def build_report():
    con = duckdb.connect(DB_PATH)
    con.execute(f"""
        CREATE OR REPLACE TABLE runs AS
        SELECT * FROM read_json_auto('{JSONL_PATH}')
    """)

    prompt_summary = con.execute("""
        SELECT prompt_id, category,
               ANY_VALUE(semantic_score) AS semantic,
               ANY_VALUE(structural_score) AS structural,
               ANY_VALUE(behavioral_score) AS behavioral,
               ANY_VALUE(classification) AS classification,
               COUNT(*) AS num_models
        FROM runs
        GROUP BY prompt_id, category
        ORDER BY category, prompt_id
    """).fetchall()

    category_summary = con.execute("""
        SELECT category,
               ROUND(AVG(semantic_score), 3) AS avg_semantic,
               ROUND(AVG(structural_score), 3) AS avg_structural,
               ROUND(AVG(behavioral_score), 3) AS avg_behavioral
        FROM runs
        GROUP BY category
        ORDER BY category
    """).fetchall()

    con.close()

    rows_html = ""
    for prompt_id, category, sem, struct, beh, classification, n in prompt_summary:
        rows_html += f"""
        <tr>
          <td>{html_lib.escape(prompt_id)}</td>
          <td>{html_lib.escape(category)}</td>
          <td style="background:{score_to_color(sem)}">{sem:.3f}</td>
          <td style="background:{score_to_color(struct)}">{struct:.3f}</td>
          <td style="background:{score_to_color(beh)}">{beh:.3f}</td>
          <td><span class="badge {classification}">{classification.replace('_',' ')}</span></td>
          <td>{n}</td>
        </tr>"""

    cat_rows_html = ""
    for category, sem, struct, beh in category_summary:
        cat_rows_html += f"""
        <tr>
          <td>{html_lib.escape(category)}</td>
          <td style="background:{score_to_color(sem)}">{sem:.3f}</td>
          <td style="background:{score_to_color(struct)}">{struct:.3f}</td>
          <td style="background:{score_to_color(beh)}">{beh:.3f}</td>
        </tr>"""

    html_doc = f"""<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Cross-Model Consistency Auditor - Report</title>
<style>
  body {{ font-family: -apple-system, Segoe UI, Arial, sans-serif; background:#0f1115; color:#e6e6e6; padding: 32px; }}
  h1 {{ font-size: 24px; }}
  h2 {{ font-size: 18px; margin-top: 40px; color:#9ecbff; }}
  table {{ border-collapse: collapse; width: 100%; margin-top: 12px; }}
  th, td {{ padding: 10px 14px; text-align: left; border-bottom: 1px solid #2a2d34; color:#111; }}
  th {{ background:#1c1f26; color:#e6e6e6; }}
  td:first-child, td:nth-child(2) {{ background:#1c1f26; color:#e6e6e6; }}
  .badge {{ padding: 3px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; color:#111;}}
  .portable {{ background:#4ade80; }}
  .cosmetically_fragile {{ background:#fde047; }}
  .structurally_fragile {{ background:#fb923c; }}
  .behaviorally_fragile {{ background:#f87171; }}
  .note {{ color:#9aa0ab; font-size: 13px; margin-top: 8px; }}
</style>
</head>
<body>
  <h1>Cross-Model Consistency Auditor</h1>
  <p class="note">Divergence scores range 0 (fully consistent across models) to 1 (fully inconsistent). Lower is better.</p>

  <h2>Per-Prompt Results</h2>
  <table>
    <tr><th>Prompt ID</th><th>Category</th><th>Semantic</th><th>Structural</th><th>Behavioral</th><th>Classification</th><th># Models</th></tr>
    {rows_html}
  </table>

  <h2>Average by Category</h2>
  <table>
    <tr><th>Category</th><th>Avg Semantic</th><th>Avg Structural</th><th>Avg Behavioral</th></tr>
    {cat_rows_html}
  </table>
</body>
</html>"""

    with open(OUTPUT_PATH, "w", encoding="utf-8") as f:
        f.write(html_doc)

    print(f"Report written to {OUTPUT_PATH}")


if __name__ == "__main__":
    build_report()