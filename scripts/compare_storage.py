"""Reproduce the old/new storage benchmark with identical current fixtures."""
from pathlib import Path
import shutil
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
output = root / "docs/benchmarks"
output.mkdir(exist_ok=True)
baseline = "2184118abb9d09846170fc2eaeef16df54eab53f"
command = ["go", "test", "./internal/store", "-run", "^$", "-bench", "BenchmarkDataset", "-benchmem", "-benchtime=10x", "-count=3"]

with tempfile.TemporaryDirectory(prefix="career-quest-baseline-") as temporary:
    before = Path(temporary).resolve()
    (before / "internal/store/migrations").mkdir(parents=True)
    for file in ("go.mod", "go.sum"):
        shutil.copy2(root / file, before / file)
    for module in ("domain", "demo"):
        shutil.copytree(root / "internal" / module, before / "internal" / module)
    for file in ("internal/store/store.go", "internal/store/migrations/001_initial.sql"):
        source = subprocess.check_output(["git", "show", f"{baseline}:{file}"], cwd=root)
        (before / file).write_bytes(source)
    shutil.copy2(root / "internal/store/benchmark_test.go", before / "internal/store/benchmark_test.go")
    for name, directory in (("before", before), ("after", root)):
        result = subprocess.run(command, cwd=directory, text=True, capture_output=True, check=True)
        (output / f"{name}.txt").write_text(result.stdout, encoding="utf-8")
        print(result.stdout)

print("Same current domain types, exercise content, scaled fixtures and benchmark; only storage implementation differs.")
