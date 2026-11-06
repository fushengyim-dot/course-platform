import unittest

from check_review_records import validate_record


VALID_RECORD = """- 记录编号：REV-001
- 对应 PR / 变更：PR #12
- 被审提交（40位 SHA）：`0123456789abcdef0123456789abcdef01234567`
- 审核人及角色：维护者自审
- 检查方式与命令：读取差异并运行指定测试
- 检查范围（文件/函数/场景）：handler route
- 检查日期：2026-11-06

## 第一层：七类代码核查
|接口入口|检查|||
|业务逻辑|检查|||
|数据库操作|检查|||
|公共函数|检查|||
|错误处理|检查|||
|安全与隐私|检查|||
|性能|检查|||
## 第二层：七个评审维度
|需求覆盖|检查|||
|一致性|检查|||
|失败路径|检查|||
|查询性能风险|检查|||
|重复代码|检查|||
|安全漏洞|检查|||
|高风险场景|检查|||
## 发现与处理
## 总结
- 总评：可用
- 未验证/不适用及理由：无
"""


class ReviewRecordTests(unittest.TestCase):
    def test_accepts_complete_record(self):
        self.assertEqual(validate_record(VALID_RECORD), [])

    def test_rejects_missing_field_and_section(self):
        result = validate_record(VALID_RECORD.replace("- 检查方式与命令：读取差异并运行指定测试\n", "").replace("## 第二层：七个评审维度\n", ""))
        self.assertTrue(any("检查方式与命令" in error for error in result))
        self.assertTrue(any("第二层" in error for error in result))

    def test_rejects_invalid_revision(self):
        result = validate_record(VALID_RECORD.replace("0123456789abcdef0123456789abcdef01234567", "not-a-sha"))
        self.assertTrue(any("40位" in error for error in result))

    def test_rejects_unlabelled_reviewer_role(self):
        result = validate_record(VALID_RECORD.replace("维护者自审", "维护者"))
        self.assertTrue(any("审核人角色" in error for error in result))

    def test_rejects_incomplete_review_tables(self):
        result = validate_record(VALID_RECORD.replace("|数据库操作|检查|||\n", ""))
        self.assertTrue(any("数据库操作" in error for error in result))


if __name__ == "__main__":
    unittest.main()
