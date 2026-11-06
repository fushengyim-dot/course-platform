#!/usr/bin/env python3
"""Require a structurally complete review record in each main-targeting PR."""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from pathlib import Path


REQUIRED_FIELDS = (
    "记录编号",
    "对应 PR / 变更",
    "被审提交（40位 SHA）",
    "审核人及角色",
    "检查方式与命令",
    "检查范围（文件/函数/场景）",
    "检查日期",
    "总评",
    "未验证/不适用及理由",
)
CATEGORY_ROWS = (
    "接口入口",
    "业务逻辑",
    "数据库操作",
    "公共函数",
    "错误处理",
    "安全与隐私",
    "性能",
)
DIMENSION_ROWS = (
    "需求覆盖",
    "一致性",
    "失败路径",
    "查询性能风险",
    "重复代码",
    "安全漏洞",
    "高风险场景",
)
REVIEW_PATH = re.compile(r"^docs/reviews/review-[^/]+\.md$")
SHA_PATTERN = re.compile(r"^[0-9a-fA-F]{40}$")


def validate_record(text: str) -> list[str]:
    errors: list[str] = []
    for field in REQUIRED_FIELDS:
        match = re.search(rf"^-\s*{re.escape(field)}[:：]\s*(.+?)\s*$", text, re.MULTILINE)
        if match is None or match.group(1) in {"", "待填写", "TODO"}:
            errors.append(f"缺少或未填写字段：{field}")

    revision = re.search(r"^-\s*被审提交（40位 SHA）[:：]\s*`?([0-9a-fA-F]{40})`?\s*$", text, re.MULTILINE)
    if revision is None or not SHA_PATTERN.fullmatch(revision.group(1)):
        errors.append("被审提交必须是40位十六进制提交SHA")

    role = re.search(r"^-\s*审核人及角色[:：]\s*(.+)$", text, re.MULTILINE)
    if role is not None and not any(label in role.group(1) for label in ("自审", "自动化发现", "独立同伴复核")):
        errors.append("审核人角色必须明确标为自审、自动化发现或独立同伴复核")

    for heading in ("第一层：七类代码核查", "第二层：七个评审维度", "发现与处理", "总结"):
        if f"## {heading}" not in text:
            errors.append(f"缺少章节：{heading}")
    for label in CATEGORY_ROWS + DIMENSION_ROWS:
        if f"|{label}|" not in text:
            errors.append(f"缺少核查表行：{label}")
    return errors


def git_changed_files(base: str, head: str) -> list[str]:
    result = subprocess.run(
        ["git", "diff", "--name-only", "--diff-filter=ACMRT", "-z", base, head],
        check=True,
        capture_output=True,
    )
    return [path.decode("utf-8", errors="strict") for path in result.stdout.split(b"\0") if path]


def validate_pull_request(base: str, head: str) -> list[str]:
    changed = git_changed_files(base, head)
    reports = [Path(path) for path in changed if REVIEW_PATH.fullmatch(path.replace("\\", "/"))]
    if not reports:
        return ["PR差异中未新增或更新 docs/reviews/review-*.md 评审记录"]

    errors: list[str] = []
    valid_record_count = 0
    for report in reports:
        try:
            text = report.read_text(encoding="utf-8")
        except (OSError, UnicodeError) as exc:
            errors.append(f"无法读取 {report.as_posix()}: {exc}")
            continue

        record_errors = validate_record(text)
        revision = re.search(r"^-\s*被审提交（40位 SHA）[:：]\s*`?([0-9a-fA-F]{40})`?\s*$", text, re.MULTILINE)
        if revision is not None:
            checked = subprocess.run(
                ["git", "cat-file", "-e", f"{revision.group(1)}^{{commit}}"],
                capture_output=True,
            )
            if checked.returncode != 0:
                record_errors.append("被审SHA不是当前仓库可读取的提交")
        if record_errors:
            errors.extend(f"{report.as_posix()}: {error}" for error in record_errors)
        else:
            valid_record_count += 1

    if valid_record_count == 0 and not errors:
        errors.append("PR差异中没有结构完整且引用有效提交的评审记录")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base", required=True, help="pull request base commit")
    parser.add_argument("--head", required=True, help="pull request head commit")
    args = parser.parse_args()

    try:
        errors = validate_pull_request(args.base, args.head)
    except subprocess.CalledProcessError as exc:
        print(f"review record check failed: git command exited {exc.returncode}", file=sys.stderr)
        return 2

    if errors:
        for error in errors:
            print(f"FAIL: {error}", file=sys.stderr)
        return 1
    print("review record check: PASS")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
