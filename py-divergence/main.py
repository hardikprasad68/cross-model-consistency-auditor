import json
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer, util

app = FastAPI()
model = SentenceTransformer('all-MiniLM-L6-v2')


def semantic_divergence(outputs: list[str]) -> float:
    embeddings = model.encode(outputs)
    sims = util.cos_sim(embeddings, embeddings)
    n = len(outputs)
    total, count = 0.0, 0
    for i in range(n):
        for j in range(i + 1, n):
            total += 1 - sims[i][j].item()
            count += 1
    return total / count if count else 0.0


def structural_divergence(outputs: list[str], expected_format: str = "text") -> float:
    if expected_format == "json":
        valid_count = 0
        for o in outputs:
            cleaned = o.strip()
            if cleaned.startswith("```"):
                cleaned = cleaned.strip("`")
                if cleaned.lower().startswith("json"):
                    cleaned = cleaned[4:].strip()
            try:
                json.loads(cleaned)
                valid_count += 1
            except (json.JSONDecodeError, ValueError):
                pass
        return 1 - (valid_count / len(outputs))

    lengths = [len(o.split()) for o in outputs]
    if not lengths or max(lengths) == 0:
        return 0.0
    return (max(lengths) - min(lengths)) / max(lengths)


REFUSAL_MARKERS = [
    "i cannot", "i can't", "i'm not able to", "i am not able to",
    "as an ai", "i must decline", "i won't be able to",
    "i'm unable to", "i am unable to", "i don't feel comfortable",
]

def behavioral_divergence(outputs: list[str]) -> float:
    flags = [any(m in o.lower() for m in REFUSAL_MARKERS) for o in outputs]
    if all(flags) or not any(flags):
        return 0.0
    return sum(flags) / len(flags)


class ScoreRequest(BaseModel):
    outputs: list[str]
    expected_format: str = "text"


@app.post("/score")
def score(req: ScoreRequest):
    return {
        "semantic": semantic_divergence(req.outputs),
        "structural": structural_divergence(req.outputs, req.expected_format),
        "behavioral": behavioral_divergence(req.outputs),
    }