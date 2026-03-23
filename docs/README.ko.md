<!-- AGENT NOTE: 이 파일을 수정하면 README.md (영어)와 docs/README.ja.md (일본어)도 함께 업데이트하세요. -->

<p align="center">
  <img src="https://img.shields.io/github/v/release/neuradex/blindenv?style=flat-square&color=blue" alt="release" />
  <img src="https://img.shields.io/badge/Claude_Code-plugin-blueviolet?style=flat-square" alt="Claude Code plugin" />
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="license" />
</p>

<h1 align="center">blindenv</h1>

<p align="center">
  <strong>AI 코딩 에이전트를 위한 시크릿 격리.</strong>
  <br>
  에이전트가 시크릿을 사용하되, 절대 보지 못하게.
</p>

<p align="center">
  <a href="#기능">기능</a> ·
  <a href="#설치">설치</a> ·
  <a href="#빠른-시작">빠른 시작</a> ·
  <a href="#동작-원리">동작 원리</a> ·
  <a href="#설정">설정</a> ·
  <a href="#cli-레퍼런스">CLI 레퍼런스</a>
</p>

<p align="center">
  <a href="../README.md">English</a> ·
  <strong>한국어</strong> ·
  <a href="./README.ja.md">日本語</a>
</p>

---

## 기능

blindenv는 AI 에이전트가 API 키, 데이터베이스 인증정보, 토큰을 **사용**하되 절대 **보지 못하게** 합니다.

- **시크릿 주입** — 격리된 서브프로세스에서 `$VAR` 참조를 해석합니다. 에이전트가 `$API_KEY`를 쓰면, 실제 값은 뒤에서 주입됩니다.
- **출력 마스킹** — 모든 stdout/stderr를 스캔하여 시크릿 값을 `[REDACTED]`로 치환한 뒤 에이전트에게 전달합니다.
- **파일 차단** — 에이전트가 `.env` 파일이나 인증정보 파일을 읽거나 검색하거나 수정하는 것을 차단합니다.
- **설정 보호** — 에이전트가 `blindenv.yml`을 수정할 수 없습니다. 규칙은 변조 불가능합니다.

```
에이전트 작성:    curl -H "Authorization: $API_KEY" https://api.example.com
                         ↓
blindenv 프록시:  서브프로세스 환경에 실제 값 주입
                         ↓
에이전트 수신:    {"result": "ok", "token": "[REDACTED]"}
```

---

## 설치

### Claude Code 플러그인 (권장)

```bash
/plugin marketplace add neuradex/blindenv
/plugin install blindenv@blindenv
```

끝입니다. 다음 세션 시작 시 플랫폼(macOS/Linux/Windows, amd64/arm64)에 맞는 바이너리가 [GitHub Releases](https://github.com/neuradex/blindenv/releases)에서 자동 다운로드됩니다.

프로젝트 루트에 `blindenv.yml`을 생성하세요:

```yaml
secret_files:
  - .env
```

완료. 모든 Bash 명령이 blindenv를 통해 실행되고, 시크릿 파일은 모든 에이전트 도구에서 차단됩니다.

### 소스에서 빌드

```bash
go install github.com/neuradex/blindenv@latest
```

### 플랫폼 지원

| 플랫폼 | 아키텍처 | |
|--------|---------|---|
| macOS | Apple Silicon (arm64) | 지원 |
| macOS | Intel (amd64) | 지원 |
| Linux | x86_64 (amd64) | 지원 |
| Linux | ARM (arm64) | 지원 |
| Windows | x86_64 (amd64) | 지원 |
| Windows | ARM64 | 지원 |

---

## 빠른 시작

`blindenv.yml`이 있으면 `.env` 파일의 모든 키-값 쌍이 자동으로:

| | 동작 |
|---|---|
| **주입** | `blindenv run`을 통해 명령에서 `$VAR`로 시크릿 값 사용 가능 |
| **마스킹** | 시크릿 값이 포함된 출력 → `[REDACTED]` |
| **차단** | 에이전트가 시크릿 파일을 Read, Grep, Edit, Write 불가 |

```bash
# .env 내용: API_KEY=sk-a1b2c3d4
blindenv run 'curl -H "Authorization: Bearer $API_KEY" https://api.example.com'
# → {"result": "ok", "key": "[REDACTED]"}
```

Claude Code 플러그인으로 사용 시에는 `blindenv run`을 직접 쓸 필요 없이 — 훅이 Bash 명령을 자동으로 래핑합니다.

---

## 동작 원리

```
┌─────────────────────────────────────────────────────┐
│  에이전트 컨텍스트 (시크릿 없음)                       │
│                                                     │
│  "curl -H 'Authorization: $API_KEY' example.com"   │
│          │                                          │
│          ▼                                          │
│  ┌─────────────────────────────────────┐            │
│  │  blindenv 프록시                     │            │
│  │  ┌──────────────┐                  │            │
│  │  │ 해석          │ API_KEY=sk-a1b2 │            │
│  │  │ 격리          │ 서브프로세스만    │            │
│  │  │ 실행          │ 실제 curl 실행   │            │
│  │  │ 마스킹        │ sk-a1b2→[REDACTED] │         │
│  │  └──────────────┘                  │            │
│  └─────────────────────────────────────┘            │
│          │                                          │
│          ▼                                          │
│  {"result": "ok", "token": "[REDACTED]"}            │
└─────────────────────────────────────────────────────┘
```

### 방어 계층

| # | 계층 | 역할 |
|---|------|------|
| 1 | **서브프로세스 격리** | 시크릿은 서브프로세스 환경에만 존재 — 에이전트 컨텍스트에는 절대 노출되지 않음 |
| 2 | **출력 마스킹** | stdout/stderr를 스캔하여 시크릿 값을 `[REDACTED]`로 치환 |
| 3 | **파일 차단** | `secret_files`에 등록된 파일에 대해 Read, Grep, Edit, Write 차단 |
| 4 | **설정 보호** | 에이전트가 `blindenv.yml`을 수정할 수 없음 — 변조 불가 |

### Claude Code 훅

플러그인 설치 시, 5개의 PreToolUse 훅이 모든 에이전트 동작을 감시합니다:

```
┌─ blindenv.yml ──────────────────────────────────────┐
│                                                      │
│  Bash 훅           Read/Grep 훅      Edit/Write 훅   │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │ 명령 래핑     │  │ 시크릿 파일   │  │ 시크릿     │ │
│  │ → blindenv   │  │ 접근 차단     │  │ 파일 +     │ │
│  │   run '...'  │  │              │  │ 설정 차단   │ │
│  │ 시크릿 주입   │  │              │  │            │ │
│  │ 출력 마스킹   │  │              │  │            │ │
│  └──────────────┘  └──────────────┘  └────────────┘ │
│                                                      │
└──────────────────────────────────────────────────────┘
```

| 도구 | 훅 | 동작 |
|------|-----|------|
| **Bash** | `blindenv hook cc bash` | 명령을 `blindenv run '...'`으로 래핑 — 시크릿 주입, 출력 마스킹 |
| **Read** | `blindenv hook cc read` | 시크릿 파일 읽기 차단 |
| **Grep** | `blindenv hook cc grep` | 시크릿 파일 검색 차단 |
| **Edit** | `blindenv hook cc guard-file` | 시크릿 파일 및 `blindenv.yml` 수정 차단 |
| **Write** | `blindenv hook cc guard-file` | 시크릿 파일 및 `blindenv.yml` 쓰기 차단 |

훅은 **exit 2** (차단 에러)를 사용하며, 모든 Claude Code 권한 모드에서 동작합니다. 제안이 아니라 — 구조적 게이트입니다.

---

## 설정

```yaml
# blindenv.yml

secret_files:        # .env 파일 — 자동 파싱, 에이전트 접근 차단
  - .env
  - .env.local
  - ~/.aws/credentials

inject:              # 호스트 프로세스의 환경변수 — 주입 + 마스킹
  - CI_TOKEN
  - DEPLOY_KEY

passthrough:         # 비밀이 아닌 변수 — 명시적 허용목록 (엄격 모드)
  - PATH
  - HOME
  - LANG
```

| 필드 | 용도 | 사용 시점 |
|------|------|----------|
| `secret_files` | `.env` 파일 파싱, 값 주입, 파일 접근 차단 | **항상** — 핵심 메커니즘 |
| `inject` | 호스트 프로세스에서 환경변수 가져오기 | CI/CD 시크릿, 파일에 없는 변수 |
| `passthrough` | 비밀이 아닌 변수의 엄격한 허용목록 | 높은 보안이 필요한 환경 |

**`secret_files`만으로도 충분합니다.** `inject`는 호스트 프로세스에는 있지만 파일에는 없는 환경변수(예: CI 시크릿)를 위한 것입니다.

`passthrough`가 설정되면, 서브프로세스는 해당 변수와 주입된 시크릿만 받습니다(엄격 모드). 설정하지 않으면 호스트 환경 전체를 상속합니다(허용 모드).

설정 파일은 `cwd`에서 `/`까지 상위로 탐색한 뒤 `~/.blindenv.yml`도 확인합니다. 가장 가까운 파일이 적용됩니다 — `.gitignore`와 같은 방식입니다.

---

## CLI 레퍼런스

```
blindenv run '<command>'              시크릿 격리 + 출력 마스킹으로 실행
blindenv check-file <path>            파일 차단 여부 확인 (exit 2 = 차단됨)
blindenv has-config                   시크릿 설정이 있으면 exit 0, 없으면 1
blindenv hook cc <hook>               Claude Code PreToolUse 훅
                                       bash | read | grep | guard-file
```

---

## 왜 .env를 그냥 쓰면 안 되나요?

AI 에이전트가 API 키가 필요합니다. 그래서 채팅에 붙여넣거나, `.env`에 넣고 에이전트가 읽게 합니다.

어느 쪽이든, 시크릿이 에이전트의 컨텍스트에 들어갑니다 — 프롬프트 인젝션 한 번이면 유출됩니다. 에이전트가 악의적일 필요도 없습니다. 속으면 되는 겁니다.

blindenv는 이것을 구조적으로 해결합니다. 프롬프트가 아니라. 신뢰가 아니라. 격리로.

---

## 라이선스

MIT

---

<p align="center">
  <strong>에이전트에게 필요한 건 당신의 키가 아닙니다. 키가 여는 것입니다.</strong>
  <br>
  blindenv는 노출 없이 접근을 제공합니다.
</p>
