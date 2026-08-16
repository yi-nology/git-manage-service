# 本地补丁副本（fork of studyzy/gongfeng-sdk-go v0.6.0）

上游 v0.6.0 的 `Time.UnmarshalJSON` 不支持工蜂 API 实际返回的
无冒号时区偏移格式（`2026-05-06T06:41:07+0000`），导致
git-platform-sdk 的 tencent_code ListRepos 解析失败。

本副本仅修改 types.go：在时间布局列表中追加
`2006-01-02T15:04:05-0700` 与 `2006-01-02T15:04:05.000-0700`。

通过 go.mod 的 replace 指令引用：
    replace github.com/studyzy/gongfeng-sdk-go => ./third_party/gongfeng-sdk-go

上游发布修复版本后可删除本目录及 replace 指令。
