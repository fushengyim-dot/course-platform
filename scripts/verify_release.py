"""Build and probe isolated Git revisions; retain verification volumes."""
import argparse
import io
import json
import os
from pathlib import Path
import secrets
import subprocess
import tarfile
import time
import urllib.request


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--ref", default="HEAD")
    parser.add_argument("--previous")
    parser.add_argument("--port", type=int, default=18772)
    args = parser.parse_args()
    if not 1024 <= args.port <= 65535:
        parser.error("port must be between 1024 and 65535")
    root = Path(__file__).resolve().parents[1]
    work = root / ".verification" / (time.strftime("%Y%m%d-%H%M%S") + "-" + secrets.token_hex(3))
    work.mkdir(parents=True)
    project = "verify-" + secrets.token_hex(5)
    password, root_password = secrets.token_hex(24), secrets.token_hex(24)
    env = dict(os.environ)
    env.update(COMPOSE_PROJECT_NAME=project, WEB_PORT=str(args.port), HTTP_ADDR=":8080",
               MYSQL_HOST="mysql", MYSQL_PORT="3306", MYSQL_DATABASE="verification",
               MYSQL_USER="verification", MYSQL_PASSWORD=password, MYSQL_ROOT_PASSWORD=root_password,
               QDRANT_URL="http://qdrant:6333", API_UPSTREAM="api:8080")
    records = []
    refs = [args.previous, args.ref, args.previous] if args.previous else [args.ref]
    report = {"project": project, "port": args.port, "runs": records, "passed": False}
    try:
        for index, ref in enumerate(refs):
            sha = subprocess.check_output(["git", "rev-parse", "--verify", ref + "^{commit}"], cwd=root, text=True).strip()
            source = work / str(index)
            source.mkdir()
            archive = subprocess.check_output(["git", "archive", sha], cwd=root)
            with tarfile.open(fileobj=io.BytesIO(archive)) as tar:
                tar.extractall(source, filter="data")
            env["APP_VERSION"] = sha[:12]
            command = ["docker", "compose", "--project-name", project]
            record = {"ref": ref, "commit": sha, "passed": False}
            records.append(record)
            start = time.monotonic()
            try:
                run = subprocess.run(command + ["up", "--build", "--detach", "--wait", "--wait-timeout", "180"], cwd=source, env=env, capture_output=True, text=True, timeout=900)
                output = (run.stdout + run.stderr).replace(password, "***").replace(root_password, "***")
                (work / (str(index) + "-compose.log")).write_text(output, encoding="utf-8")
                if run.returncode:
                    raise RuntimeError("Compose startup failed; see local verification log")
                with urllib.request.urlopen("http://127.0.0.1:" + str(args.port) + "/health", timeout=5) as response:
                    body = json.load(response)
                    if response.status != 200 or body.get("code") != "OK":
                        raise RuntimeError("health response failed")
                if body["data"]["version"] != sha[:12] or not all(v == "up" for v in body["data"]["dependencies"].values()):
                    raise RuntimeError("version or dependency check failed")
                with urllib.request.urlopen("http://127.0.0.1:" + str(args.port) + "/", timeout=5) as response:
                    if response.status != 200 or b"/assets/" not in response.read():
                        raise RuntimeError("page check failed")
                record.update(passed=True, health=body["data"], elapsed_seconds=round(time.monotonic()-start, 2))
                print(json.dumps(record), flush=True)
            finally:
                down = subprocess.run(command + ["down", "--timeout", "10"], cwd=source, env=env, capture_output=True, text=True, timeout=90)
                record["stopped"] = down.returncode == 0
                if down.returncode:
                    raise RuntimeError("verification cleanup failed")
        report["passed"] = True
    except Exception as error:
        report["error"] = str(error).replace(password, "***").replace(root_password, "***")
        raise
    finally:
        (work / "report.json").write_text(json.dumps(report, indent=2), encoding="utf-8")
        print("Verification report: " + str(work / "report.json"), flush=True)


if __name__ == "__main__":
    main()
