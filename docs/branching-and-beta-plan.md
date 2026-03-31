# 分支与 Beta 发布规划

本文档用于约定 `Rubin Investment` 在短期内的分支策略和版本发布方式。

## 当前阶段

项目当前仍处于功能持续迭代阶段，短期内统一采用 `beta` 版本发布，不承诺稳定版语义。

当前已知仍在持续打磨的方向包括：

- 持仓模块仍在开发中
- 量化联动推荐功能稳定性仍需优化

因此，短期版本建议统一使用：

- `v0.x.y-beta.N`

示例：

- `v0.1.0-beta.1`
- `v0.1.0-beta.2`
- `v0.2.0-beta.1`

## 推荐分支模型

### 1. `main`

用途：

- GitHub 默认展示分支
- 对外公开的主分支
- 只保留“可运行、可演示、可发布 beta”的代码

规则：

- 不直接在 `main` 上做日常开发
- 只有在准备发布新 beta 时才合并进入
- 每次进入 `main` 都应对应一个可说明的 beta 版本

### 2. `develop`

用途：

- 日常集成分支
- 当前阶段最主要的开发分支

规则：

- 功能开发先进入 `feature/*`
- 功能合并后先回到 `develop`
- 验证通过后，再从 `develop` 合并到 `main`

### 3. `feature/*`

用途：

- 单个功能或单个修复的工作分支

命名建议：

- `feature/asset-holdings`
- `feature/quant-linkage-improve`
- `feature/settings-security-tab`

规则：

- 一个分支只处理一类问题
- 完成后合并回 `develop`

### 4. `fix/*`

用途：

- 非紧急修复
- 发布前修正 beta 问题

命名建议：

- `fix/asset-unlock-flow`
- `fix/quant-recommend-timeout`

### 5. `hotfix/*`

用途：

- 已经进入 `main` 的 beta 版本出现明显问题时使用

命名建议：

- `hotfix/beta-login-regression`
- `hotfix/export-build-failure`

规则：

- 从 `main` 拉出
- 修完后同时合并回 `main` 和 `develop`

## 短期建议流程

短期内建议采用下面这套简单流程：

1. 首次提交后，将默认分支整理为 `main`
2. 从 `main` 拉出 `develop`
3. 日常开发全部走 `feature/*` 或 `fix/*`
4. 功能完成后先合并到 `develop`
5. 当 `develop` 达到可演示状态时，合并到 `main`
6. 在 `main` 上打 `beta tag`

## Beta 版本规则

### 版本号建议

在正式稳定版之前，统一使用：

- `v0.1.0-beta.N`
- `v0.2.0-beta.N`

推荐含义：

- `0.x`：表示整体仍处于早期阶段
- `y`：小范围功能整理或修复
- `beta.N`：同一轮对外测试的递增发布号

### 何时递增版本

- 新增一组可感知功能：升中间位，例如 `v0.1.0-beta.3` -> `v0.2.0-beta.1`
- 同一轮功能的小修复：只增加 `beta.N`

## 当前阶段推荐命名

如果你准备马上开始正式用 GitHub 管理，建议直接采用：

- 默认分支：`main`
- 集成分支：`develop`

第一批功能分支建议示例：

- `feature/holdings-module`
- `feature/quant-linkage-stability`
- `feature/asset-analysis-security`

## 发布说明建议

每次 beta 发布建议在 Release 或提交说明里明确写出：

- 本次新增内容
- 本次修复内容
- 当前已知问题
- 是否建议普通用户使用

当前已知问题建议持续保留：

- 持仓模块尚未完成
- 量化联动推荐存在不稳定情况

## 进入稳定版前的信号

当以下条件基本满足时，再考虑去掉 `beta`：

- 持仓模块完成可用闭环
- 量化联动推荐稳定性明显提升
- 配置、打包、导出流程趋于稳定
- README、发布说明、已知问题维护成熟
