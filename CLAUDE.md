# x-mindflow AI Coding 规范

## 📌 项目概述

**x-mindflow** 是一个企业级思维导图与知识流动管理平台，采用前后端分离架构。

### 技术栈

#### 后端
- **语言**: Go 1.24
- **Web 框架**: Gin
- **ORM**: GORM
- **缓存**: freecache (内存缓存)
- **日志**: uber-go/zap
- **数据库**: PostgreSQL 18.0

#### 前端
- **运行时**: Node.js 20
- **框架**: Vue 3 + TypeScript
- **状态管理**: Pinia
- **HTTP 客户端**: Axios
- **样式方案**: Tailwind CSS
- **图可视化**: AntV G6

#### API 协议
- HTTP + JSON
- 遵循 RESTful API 规范

#### 部署方式
- **开发环境**: Docker Compose
- **生产环境**: Kubernetes (K8s)

---

## 🔧 开发环境设置

### 前置依赖
- Go 1.24+
- Node.js 20+
- PostgreSQL 18.0+
- Docker & Docker Compose
- Make

### 一键启动开发环境

```bash
# 初始化项目（安装依赖、数据库迁移等）
make init

# 启动完整开发环境（前端 + 后端 + 数据库）
make dev

# 仅启动后端
make dev-backend

# 仅启动前端
make dev-frontend
```

### 构建项目

```bash
# 构建所有组件
make build

# 构建后端
make build-backend

# 构建前端
make build-frontend
```

---

## 📋 工作流程规范

### Issue 驱动开发流程

我们采用严格的 Issue 拆分策略，确保每个阶段产出明确的交付物：

```
需求分析 Issue
    ↓
原型设计 Issue (产出: 高保真原型 + 原型描述文档 PR)
    ↓
模块设计 Issue (产出: API 设计 + 数据模型 + 架构设计文档 PR)
    ↓
开发任务 Sub-Issues (每个模块的具体开发任务，可并行)
```

#### 1️⃣ 需求分析阶段
- **产出**: 需求分析文档 (`/docs/requirements/`)
- **内容**: 业务背景、用户故事、功能需求、非功能需求
- **提交**: PR 提交文档后关闭 Issue

#### 2️⃣ 原型设计阶段
- **产出**: 
  - 高保真原型界面（前端实现）
  - 原型描述文档 (`/docs/prototypes/`)
- **内容**: UI/UX 设计、交互流程、页面结构
- **提交**: PR 包含前端原型代码 + 描述文档

#### 3️⃣ 模块设计阶段
- **产出**: 
  - API 设计文档 (`/docs/api/`)
  - 数据模型设计 (`/docs/database/`)
  - 架构设计文档 (`/docs/architecture/`)
  - 模块设计文档 (`/docs/modules/`)
- **内容**: 
  - RESTful API 端点定义（请求/响应格式）
  - 数据库表结构（字段、索引、关系）
  - 系统架构图（组件交互、数据流）
  - 模块职责与接口定义
- **提交**: PR 提交所有设计文档

#### 4️⃣ 开发实现阶段
- **拆分**: 基于模块设计创建多个 Sub-Issues
- **并行**: 独立模块可同时开发
- **产出**: 功能代码 + 单元测试
- **提交**: 每个 Sub-Issue 对应独立 PR

---

## 📝 Git 工作流规范

### 分支命名规范

```
feature/<issue-number>-<feature-name>    # 新功能
bugfix/<issue-number>-<bug-description>  # Bug 修复
hotfix/<issue-number>-<critical-fix>     # 紧急修复
refactor/<scope>                         # 重构
docs/<doc-type>                          # 文档更新
```

**示例**:
- `feature/123-user-authentication`
- `bugfix/456-login-error`
- `hotfix/789-critical-data-leak`

### Commit Message 规范

采用 **Conventional Commits** 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

#### Type 类型
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构（不新增功能或修复 Bug）
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建工具、依赖更新等
- `ci`: CI/CD 配置

#### 示例

```bash
feat(auth): add JWT token refresh mechanism

- Implement token refresh endpoint
- Add middleware to check token expiration
- Update frontend to handle token renewal

Closes #123
```

```bash
fix(api): resolve null pointer in user query

The getUserByID function didn't handle deleted users properly,
causing a panic when accessing nil user objects.

Fixes #456
```

### Pull Request 规范

#### PR 标题格式
```
<type>(<scope>): <brief description>
```

**示例**:
- `feat(mindmap): add collaborative editing feature`
- `fix(auth): resolve session timeout issue`

#### PR 描述模板

```markdown
## 📝 变更说明
简要描述本次 PR 的目的和内容。

## 🔗 关联 Issue
Closes #<issue-number>

## ✨ 主要变更
- 变更点 1
- 变更点 2
- 变更点 3

## 🧪 测试说明
- [ ] 单元测试已通过
- [ ] 集成测试已通过
- [ ] 手动测试已完成

## 📸 截图/演示
（如果是 UI 变更，请提供截图或 GIF）

## ⚠️ 注意事项
（如有需要迁移、配置变更等，请说明）
```

#### PR Checklist
- [ ] 代码遵循项目风格规范
- [ ] 已添加必要的单元测试
- [ ] 测试覆盖率符合要求
- [ ] 已更新相关文档
- [ ] 已运行 lint 和格式化工具
- [ ] 已通过所有 CI 检查
- [ ] 正确填写 Assignee 和 Labels

### Issue 规范

#### Issue 标题格式
```
[<类型>] <简要描述>
```

**类型**:
- `需求分析`
- `原型设计`
- `模块设计`
- `功能开发`
- `Bug`
- `优化`
- `文档`

**示例**:
- `[需求分析] 用户协作编辑功能需求分析`
- `[原型设计] 思维导图编辑器界面原型`
- `[Bug] 登录后 Token 未正确存储`

#### Issue 必填项
- **Assignee**: 指定负责人
- **Labels**: 至少一个标签（如 `enhancement`, `bug`, `documentation`）
- **Milestone**: 关联里程碑（如适用）
- **Description**: 清晰描述问题或需求

#### Issue 描述模板

**需求/设计类 Issue**:
```markdown
## 背景
说明为什么需要这个功能或设计。

## 目标
明确本 Issue 的交付目标。

## 详细描述
详细说明需求或设计内容。

## 验收标准
- [ ] 标准 1
- [ ] 标准 2

## 参考资料
（如有相关文档或链接）
```

**Bug Issue**:
```markdown
## Bug 描述
简要描述 Bug 现象。

## 复现步骤
1. 步骤 1
2. 步骤 2
3. 观察到的错误

## 预期行为
应该发生什么。

## 实际行为
实际发生了什么。

## 环境信息
- OS: 
- Browser: 
- Version: 

## 相关日志/截图
（如有）
```

---

## 🎨 前端代码风格规范

### 文件组织

```
src/
├── assets/          # 静态资源（图片、字体等）
├── components/      # 通用组件
│   ├── common/     # 基础组件（Button, Input 等）
│   └── business/   # 业务组件
├── views/          # 页面组件
├── router/         # 路由配置
├── stores/         # Pinia 状态管理
├── api/            # API 接口封装
├── utils/          # 工具函数
├── composables/    # 组合式函数（Composition API）
├── types/          # TypeScript 类型定义
└── styles/         # 全局样式
```

### 命名规范

#### 文件命名
- **组件文件**: PascalCase（如 `UserProfile.vue`）
- **工具文件**: camelCase（如 `formatDate.ts`）
- **Store 文件**: camelCase + Store 后缀（如 `userStore.ts`）
- **类型文件**: PascalCase（如 `User.ts`）

#### 变量命名
```typescript
// 常量：UPPER_SNAKE_CASE
const MAX_RETRY_COUNT = 3;

// 普通变量/函数：camelCase
const userName = 'Alice';
function getUserInfo() {}

// 类型/接口：PascalCase
interface UserProfile {}
type ApiResponse = {};

// 组件名：PascalCase
const UserCard = defineComponent({});
```

### Vue 组件风格

#### 组件结构顺序
```vue
<script setup lang="ts">
// 1. 导入依赖
import { ref, computed } from 'vue';
import { useUserStore } from '@/stores/userStore';

// 2. 类型定义
interface Props {
  userId: string;
  userName: string;
}

// 3. Props 和 Emits
const props = defineProps<Props>();
const emit = defineEmits<{
  (e: 'update', id: string): void;
}>();

// 4. 组合式函数 (Composables)
const userStore = useUserStore();

// 5. 响应式数据
const isLoading = ref(false);

// 6. 计算属性
const displayName = computed(() => props.userName);

// 7. 方法
function handleUpdate() {
  emit('update', props.userId);
}

// 8. 生命周期钩子
onMounted(() => {
  // ...
});
</script>

<template>
  <!-- 单根元素，使用语义化标签 -->
  <div class="user-profile">
    <h2 class="text-xl font-bold">{{ displayName }}</h2>
    <button
      class="btn-primary"
      @click="handleUpdate"
    >
      Update
    </button>
  </div>
</template>

<style scoped>
/* 仅在必要时使用 <style>，优先使用 Tailwind */
.user-profile {
  /* ... */
}
</style>
```

#### Composition API 优先
- 使用 `<script setup>` 语法
- 优先使用组合式函数 (Composables) 复用逻辑
- 避免使用 Options API

#### Props 定义
```typescript
// ✅ 推荐：使用 TypeScript 接口
interface Props {
  userId: string;
  isActive?: boolean;  // 可选属性
  count?: number;
}

const props = withDefaults(defineProps<Props>(), {
  isActive: false,
  count: 0,
});

// ❌ 避免：运行时 props 定义
const props = defineProps({
  userId: String,
  isActive: Boolean,
});
```

### Tailwind CSS 使用规范

#### 类名顺序约定
```vue
<template>
  <!-- 顺序：布局 → 盒模型 → 排版 → 视觉 → 交互 -->
  <div
    class="
      flex items-center justify-between
      w-full h-16 p-4 m-2
      text-lg font-semibold text-gray-800
      bg-white rounded-lg shadow-md
      hover:bg-gray-50 cursor-pointer
    "
  >
    Content
  </div>
</template>
```

#### 复杂样式抽取
```typescript
// utils/styles.ts
export const buttonStyles = {
  base: 'px-4 py-2 rounded font-medium transition-colors',
  primary: 'bg-blue-600 text-white hover:bg-blue-700',
  secondary: 'bg-gray-200 text-gray-800 hover:bg-gray-300',
};
```

```vue
<template>
  <button :class="[buttonStyles.base, buttonStyles.primary]">
    Submit
  </button>
</template>
```

#### 响应式设计
```vue
<template>
  <div class="
    grid grid-cols-1
    md:grid-cols-2
    lg:grid-cols-3
    xl:grid-cols-4
    gap-4
  ">
    <!-- 移动端 1 列，平板 2 列，桌面 3 列，大屏 4 列 -->
  </div>
</template>
```

### TypeScript 规范

#### 类型优先，避免 any
```typescript
// ✅ 推荐
interface User {
  id: string;
  name: string;
  email: string;
}

function getUser(id: string): Promise<User> {
  return api.get<User>(`/users/${id}`);
}

// ❌ 避免
function getUser(id: any): any {
  return api.get(`/users/${id}`);
}
```

#### API 接口类型定义
```typescript
// types/api.ts
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

// api/user.ts
export async function getUserInfo(id: string): Promise<ApiResponse<User>> {
  return request.get(`/users/${id}`);
}
```

### Pinia Store 规范

```typescript
// stores/userStore.ts
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { User } from '@/types/user';

export const useUserStore = defineStore('user', () => {
  // State
  const currentUser = ref<User | null>(null);
  const isAuthenticated = ref(false);

  // Getters
  const userName = computed(() => currentUser.value?.name ?? 'Guest');

  // Actions
  async function login(email: string, password: string) {
    try {
      const response = await api.login(email, password);
      currentUser.value = response.data;
      isAuthenticated.value = true;
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  }

  function logout() {
    currentUser.value = null;
    isAuthenticated.value = false;
  }

  return {
    // State
    currentUser,
    isAuthenticated,
    // Getters
    userName,
    // Actions
    login,
    logout,
  };
});
```

### Axios 请求封装

```typescript
// utils/request.ts
import axios from 'axios';
import type { ApiResponse } from '@/types/api';

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000,
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

// 响应拦截器
request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    // 统一错误处理
    console.error('Request failed:', error);
    return Promise.reject(error);
  },
);

export default request;
```

---

## 🔧 后端代码风格规范

### 项目结构

```
backend/
├── cmd/
│   └── server/          # 应用入口
│       └── main.go
├── internal/            # 私有代码
│   ├── api/            # API 处理器
│   │   ├── handler/    # HTTP 处理函数
│   │   └── middleware/ # 中间件
│   ├── service/        # 业务逻辑
│   ├── repository/     # 数据访问层
│   ├── model/          # 数据模型
│   └── config/         # 配置管理
├── pkg/                # 可导出的公共库
│   └── utils/
├── migrations/         # 数据库迁移文件
├── test/              # 测试文件
└── go.mod
```

### 命名规范

```go
// 包名：小写，单数，简短
package user

// 接口：名词或形容词，以 -er 结尾
type Reader interface {}
type UserService interface {}

// 结构体：PascalCase
type UserProfile struct {}

// 方法/函数：驼峰命名，动词开头
func (u *User) GetFullName() string {}
func CreateUser() {}

// 常量：PascalCase 或 UPPER_SNAKE_CASE
const MaxRetryCount = 3
const DEFAULT_TIMEOUT = 30
```

### Gin 路由组织

```go
// internal/api/router.go
func SetupRouter(r *gin.Engine) {
    // 中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    
    // API v1 路由组
    v1 := r.Group("/api/v1")
    {
        // 用户相关
        users := v1.Group("/users")
        {
            users.GET("", handler.ListUsers)
            users.POST("", handler.CreateUser)
            users.GET("/:id", handler.GetUser)
            users.PUT("/:id", handler.UpdateUser)
            users.DELETE("/:id", handler.DeleteUser)
        }
        
        // 思维导图相关
        mindmaps := v1.Group("/mindmaps")
        mindmaps.Use(middleware.Auth()) // 认证中间件
        {
            mindmaps.GET("", handler.ListMindmaps)
            mindmaps.POST("", handler.CreateMindmap)
        }
    }
}
```

### 错误处理

```go
// pkg/errors/errors.go
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
    return e.Message
}

// 预定义错误
var (
    ErrNotFound      = &AppError{Code: 404, Message: "Resource not found"}
    ErrUnauthorized  = &AppError{Code: 401, Message: "Unauthorized"}
    ErrInternalError = &AppError{Code: 500, Message: "Internal server error"}
)

// Handler 中使用
func GetUser(c *gin.Context) {
    user, err := service.GetUserByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, ErrNotFound)
        return
    }
    c.JSON(http.StatusOK, user)
}
```

### GORM 模型定义

```go
// internal/model/user.go
type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    Username string `gorm:"uniqueIndex;not null" json:"username"`
    Email    string `gorm:"uniqueIndex;not null" json:"email"`
    Password string `gorm:"not null" json:"-"` // 不返回密码
    
    // 关联
    Mindmaps []Mindmap `gorm:"foreignKey:UserID" json:"mindmaps,omitempty"`
}

// TableName 自定义表名
func (User) TableName() string {
    return "users"
}
```

### 日志规范 (zap)

```go
// pkg/logger/logger.go
var Logger *zap.Logger

func InitLogger() {
    config := zap.NewProductionConfig()
    config.OutputPaths = []string{"stdout", "logs/app.log"}
    Logger, _ = config.Build()
}

// 使用示例
func CreateUser(user *User) error {
    logger.Logger.Info("Creating user",
        zap.String("username", user.Username),
        zap.String("email", user.Email),
    )
    
    if err := db.Create(user).Error; err != nil {
        logger.Logger.Error("Failed to create user",
            zap.Error(err),
            zap.String("username", user.Username),
        )
        return err
    }
    
    return nil
}
```

---

## 🧪 测试规范

### 测试文件组织

```
backend/
├── internal/
│   ├── service/
│   │   ├── user_service.go
│   │   └── user_service_test.go
```

```
frontend/
├── src/
│   ├── components/
│   │   ├── UserCard.vue
│   │   └── UserCard.spec.ts
```

### 后端测试（Go）

#### 表驱动测试（Table-Driven Tests）

```go
// internal/service/user_service_test.go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *User
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid user",
            input: &User{
                Username: "alice",
                Email:    "alice@example.com",
            },
            wantErr: false,
        },
        {
            name: "duplicate username",
            input: &User{
                Username: "existing_user",
                Email:    "new@example.com",
            },
            wantErr: true,
            errMsg:  "username already exists",
        },
        {
            name: "invalid email",
            input: &User{
                Username: "bob",
                Email:    "invalid-email",
            },
            wantErr: true,
            errMsg:  "invalid email format",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := userService.CreateUser(tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### 测试命令

```bash
# 运行所有测试
make test

# 运行单个包的测试
go test ./internal/service/...

# 查看覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 前端测试

#### 单元测试（Vitest）

```typescript
// src/components/UserCard.spec.ts
import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import UserCard from './UserCard.vue';

describe('UserCard', () => {
  it('renders user name correctly', () => {
    const wrapper = mount(UserCard, {
      props: {
        userName: 'Alice',
      },
    });
    
    expect(wrapper.text()).toContain('Alice');
  });

  it('emits update event on button click', async () => {
    const wrapper = mount(UserCard, {
      props: { userId: '123', userName: 'Alice' },
    });
    
    await wrapper.find('button').trigger('click');
    
    expect(wrapper.emitted('update')).toBeTruthy();
    expect(wrapper.emitted('update')![0]).toEqual(['123']);
  });
});
```

#### 测试命令

```bash
# 运行所有测试
npm test

# 监听模式
npm run test:watch

# 覆盖率报告
npm run test:coverage
```

### 测试覆盖率要求

- **核心业务逻辑**: ≥ 80%
- **工具函数**: ≥ 90%
- **API 处理器**: ≥ 70%

---

## 🚀 代码质量检查

### 后端（Go）

```bash
# Lint 检查
make lint-backend
# 或
golangci-lint run ./...

# 格式化代码
make fmt-backend
# 或
gofmt -w .
go mod tidy
```

### 前端

```bash
# Lint 检查
npm run lint

# 自动修复
npm run lint:fix

# 格式化代码
npm run format

# 类型检查
npm run type-check
```

### 提交前必须执行

```bash
# 后端
make lint-backend test-backend

# 前端
npm run lint && npm run type-check && npm test
```

---

## 🤖 AI Coding 约束

### ✅ AI 可以做的事情

1. **代码实现**: 根据明确的设计文档编写代码
2. **Bug 修复**: 修复已识别的问题
3. **代码重构**: 优化代码结构，提升可读性
4. **测试编写**: 为新代码添加单元测试
5. **文档更新**: 同步更新代码相关文档
6. **Lint 修复**: 自动修复代码风格问题

### ⛔ AI 不能做的事情

1. **未经讨论的设计决策**: 架构变更、技术栈替换等
2. **跳过文档阶段**: 必须先产出设计文档，再开始编码
3. **修改 .github/workflows**: 禁止修改 CI/CD 配置
4. **提交敏感信息**: 密钥、Token、密码等
5. **直接合并 PR**: 需要人工 Code Review

### 🔍 AI 完成任务后必须执行

```bash
# 1. 运行 Lint 检查
make lint  # 或 npm run lint

# 2. 运行测试
make test  # 或 npm test

# 3. 类型检查（前端）
npm run type-check

# 4. 确认构建成功
make build
```

### 📌 AI 工作流程

```
接收 Issue → 阅读设计文档 → 实现代码 → 编写测试 → 
运行 Lint/Test → 提交 PR → 等待 Code Review
```

**重要提醒**: 
- AI 必须严格遵循已有的设计文档，不得自行变更设计
- 每次代码提交前必须通过 Lint 和测试
- PR 描述必须清晰说明变更内容和测试情况

---

## 📚 文档管理

### 文档目录结构

```
docs/
├── requirements/       # 需求分析文档
├── prototypes/        # 原型设计文档
├── api/              # API 设计文档
├── database/         # 数据库设计文档
├── architecture/     # 架构设计文档
├── modules/          # 模块设计文档
└── development/      # 开发指南
```

### 文档命名规范

```
<issue-number>-<类型>-<简短描述>.md
```

**示例**:
- `123-requirement-user-auth.md`
- `124-prototype-mindmap-editor.md`
- `125-api-user-management.md`

### 文档更新规则

- 设计文档通过 PR 提交到 `/docs` 目录
- 开发前必须有对应的设计文档
- 文档变更同样需要 Code Review

---

## 🔒 安全规范

1. **敏感信息**: 使用环境变量，禁止硬编码
2. **密码存储**: 必须加密（bcrypt 或 argon2）
3. **SQL 注入防护**: 使用 GORM 参数化查询
4. **XSS 防护**: 前端输入验证 + 后端转义
5. **CSRF 防护**: 使用 CSRF Token
6. **API 认证**: JWT Token + Refresh Token 机制

---

## 📦 依赖管理

### 后端（Go）

```bash
# 添加依赖
go get github.com/example/package

# 更新依赖
go get -u github.com/example/package

# 清理无用依赖
go mod tidy
```

### 前端（npm）

```bash
# 添加依赖
npm install <package>

# 添加开发依赖
npm install -D <package>

# 更新依赖（谨慎操作）
npm update
```

---

## 🎯 AI 提示词约束总结

**给 AI 的核心指令**:

```
你是 x-mindflow 项目的开发助手，必须严格遵守以下规则：

1. 开发流程：
   - 必须先阅读 /docs 下的对应设计文档
   - 不得跳过设计阶段直接编码
   - 每个 PR 必须关联明确的 Issue

2. 代码规范：
   - 后端：Go 1.24 + Gin + GORM + PostgreSQL
   - 前端：Vue 3 + TypeScript + Tailwind CSS + Pinia
   - 遵循项目既定的代码风格和命名规范

3. 测试要求：
   - 每个功能必须包含单元测试
   - 使用表驱动测试风格（Go）
   - 测试覆盖率：核心逻辑 ≥ 80%

4. 质量检查：
   - 提交前运行 make lint 和 make test
   - 前端额外运行 npm run type-check
   - 确保所有检查通过

5. Git 规范：
   - Commit 格式：<type>(<scope>): <subject>
   - 分支命名：<type>/<issue-number>-<description>
   - PR 必须填写完整的描述和 Checklist

6. 禁止事项：
   - 禁止修改 .github/workflows
   - 禁止提交密钥、Token 等敏感信息
   - 禁止跳过设计文档直接编码
   - 禁止在未经讨论的情况下改变架构

7. 文档要求：
   - 设计文档提交到 /docs 对应目录
   - API 变更必须更新 API 文档
   - 数据库变更必须提供迁移脚本
```

---

## 📞 联系与反馈

如有疑问或建议，请通过 Issue 与团队沟通。

---

**最后更新**: 2025-10-24
