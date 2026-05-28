# 后端开发指南

UniRTM 具有极高的可扩展性。如果某个工具既没有 asdf 插件，也没有 GitHub Release，你可以编写一个基于 Go 语言的后端插件。其核心接口需要实现 `ResolveVersion()`、`Download()` 和 `Install()` 方法。