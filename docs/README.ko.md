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
  <a href="#보안-모드">보안 모드</a> ·
  <a href="#설정">설정</a> ·
  <a href="#cli-레퍼런스">CLI 레퍼런스</a> ·
  <a href="#시크릿-매니저-그-너머">비교</a>
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
- **파일 은닉** — 시크릿 파일은 읽기/수정 시 "파일이 존재하지 않음"을 반환하고, 검색과 파일 목록에서 소리소문 없이 제외됩니다. 에이전트는 파일이 있는지조차 모릅니다.
- **설정 보호** — 에이전트가 `blindenv.yml`을 수정할 수 없습니다. 규칙은 변조 불가능합니다.
- **내용 기반 차단** — 시크릿 파일을 복사하거나 이름을 바꿔도, 시크릿 값이 포함된 파일은 차단됩니다. 경로 우회는 통하지 않습니다.

```
에이전트 작성:    curl -H "Authorization: $API_KEY" https://api.example.com
                         ↓
blindenv 프록시:  서브프로세스 환경에 실제 값 주입
                         ↓
에이전트 수신:    {"result": "ok", "token": "[REDACTED]"}
```

### 어떤 수작을 부려도 볼 수 없습니다

에이전트가 무엇을 시도하든, 시크릿 값은 절대 볼 수 없습니다:

| 에이전트 시도 | 에이전트가 보는 것 |
|---|---|
| Read 도구로 `.env` 읽기 | `File does not exist` |
| Edit 또는 Write로 `.env` 수정 | `File does not exist` |
| `grep API_KEY .env` | 결과 없음 (시크릿 파일이 소리소문 없이 제외됨) |
| `Glob **/.env*`로 파일 탐색 | 결과 없음 (시크릿 파일이 소리소문 없이 제외됨) |
| Bash에서 `cat .env` | 서브프로세스에서 시크릿 파일 접근 불가 |
| `.env`를 `tmp.txt`로 복사 후 읽기 | `File does not exist` (내용 기반 스캔) |
| `echo $API_KEY`로 값 출력 | `[REDACTED]` |
| `blindenv.yml` 수정하여 규칙 해제 | 차단 — 설정 변조 불가 |

에이전트는 시크릿 파일을 읽을 수도, 찾을 수도, 검색할 수도, 수정할 수도, 쓸 수도 없습니다 — 에이전트가 아는 한, **그 파일들은 존재하지 않으니까요**. 그러면서도 API 호출은 성공하고, 배포는 완료되고, 서비스는 응답합니다. 에이전트는 그것을 가능하게 하는 인증 정보를 보지 못한 채 모든 기능을 사용합니다.

---

## 설치

### Claude Code 플러그인 (권장)

```bash
/plugin marketplace add neuradex/blindenv
/plugin install blindenv@blindenv
```

끝입니다. 다음 세션 시작 시 플랫폼에 맞는 바이너리가 [GitHub Releases](https://github.com/neuradex/blindenv/releases)에서 자동 다운로드되고, 프로젝트 루트에 `blindenv.yml`이 자동 생성됩니다.

`blindenv.yml`을 열어 보호할 시크릿 파일을 설정하세요:

```yaml
secret_files:
  - .env
  - .env.local
  # - ~/.aws/credentials
```

설정이 완료되면 에이전트는 이 파일들이 존재하는지조차 알 수 없습니다 — 모든 접근이 구조적으로 차단됩니다.

### 소스에서 빌드

```bash
git clone https://github.com/neuradex/blindenv.git
cd blindenv
make build      # → ./blindenv
```

또는 글로벌 설치: `go install github.com/neuradex/blindenv@latest`

개발 환경 구성과 프로젝트 구조는 [CONTRIBUTING.md](../CONTRIBUTING.md)를 참고하세요.

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

<p align="center">
  <img src="./blindenv-architecture-light.svg" alt="blindenv 방어 아키텍처" />
</p>

### 방어 계층

| # | 계층 | 역할 |
|---|------|------|
| 1 | **서브프로세스 격리** | 시크릿은 서브프로세스 환경에만 존재 — 에이전트 컨텍스트에는 절대 노출되지 않음 |
| 2 | **출력 마스킹** | stdout/stderr를 스캔하여 시크릿 값을 `[REDACTED]`로 치환 |
| 3 | **파일 은닉** | 시크릿 파일은 Read/Edit/Write 시 "존재하지 않음" 반환, Grep/Glob 결과에서 소리소문 없이 제외 |
| 4 | **설정 보호** | 에이전트가 `blindenv.yml`을 수정할 수 없음 — 변조 불가 |
| 5 | **내용 기반 차단** | 시크릿 값이 포함된 파일은 경로와 무관하게 차단 — 복사나 이름 변경으로 우회 불가 |

### Claude Code 훅

플러그인 설치 시, 6개의 PreToolUse 훅이 모든 에이전트 동작을 감시합니다:

<p align="center">
  <img src="./blindenv-hooks-light.svg" alt="Claude Code 훅 아키텍처" />
</p>

| 도구 | 훅 | 동작 |
|------|-----|------|
| **Bash** | `blindenv hook cc bash` | 명령을 `blindenv run '...'`으로 래핑 — 시크릿 주입, 출력 마스킹 |
| **Read** | `blindenv hook cc read` | 시크릿 파일 → "존재하지 않음" (blindenv 흔적 없음) |
| **Grep** | `blindenv hook cc grep` | 제외 glob 주입 — 시크릿 파일이 검색 결과에서 소리소문 없이 제외 |
| **Glob** | `blindenv hook cc glob` | 제외 패턴 주입 — 시크릿 파일이 파일 목록에서 소리소문 없이 제외 |
| **Edit** | `blindenv hook cc guard-file` | 시크릿 파일 → "존재하지 않음"; `blindenv.yml` → 명시적 차단 |
| **Write** | `blindenv hook cc guard-file` | 시크릿 파일 → "존재하지 않음"; `blindenv.yml` → 명시적 차단 |

에이전트는 시크릿 파일에 대해 "차단됨" 메시지를 절대 보지 못합니다 — 파일이 존재하지 않을 뿐입니다. 유일한 명시적 차단은 `blindenv.yml` 자체이며, 이는 에이전트가 blindenv 설치를 이미 알고 있기 때문입니다.

---

## 보안 모드

> *"존재 자체를 모르는 것이 최고의 보안이다."*

blindenv는 점점 강해지는 세 가지 보안 모드를 제공합니다:

```yaml
# blindenv.yml
mode: stealth    # block (기본값) | stealth | evacuate
```

| 모드 | 시크릿 파일 접근 | `ls`에 파일 노출? | 적합한 경우 |
|------|-----------------|------------------|-----------|
| **`block`** | 명시적 차단 + 출력 마스킹 | 예 | 대부분의 프로젝트 — 차단 시 명확한 피드백 |
| **`stealth`** | 파일이 존재하지 않는 것처럼 보임 | 예 (디스크에는 존재) | 에이전트가 시크릿 존재를 몰라야 할 때 |
| **`evacuate`** | 파일이 존재하지 않는 것처럼 보임 | **아니오** (물리적 삭제) | 최고 보안 — `ls`로도 아무것도 안 보임 |

### `block` — 차단과 마스킹 (기본값)

에이전트가 시크릿 파일에 접근하면 "접근 거부" 메시지를 봅니다. 출력은 여전히 마스킹됩니다. 에이전트는 파일이 있다는 것은 알지만 접근할 수 없습니다. 대부분의 프로젝트에서 가장 안전한 기본값입니다.

### `stealth` — 파일이 존재하지 않음

시크릿 파일이 아예 존재하지 않는 것처럼 보입니다. Read는 "파일 없음"을 반환하고, 검색에서 소리소문 없이 제외됩니다. 에이전트는 파일이 있는지 알 방법이 없습니다 — Bash에서 `ls`를 실행하지 않는 한.

### `evacuate` — 완전한 은닉

가장 강력한 모드. 세션 시작 시 시크릿 파일을 안전한 캐시(`~/.cache/blindenv/`)로 옮기고 디스크에서 물리적으로 삭제합니다. Bash에서 `ls`, `find`, `tree`를 실행해도 아무것도 나오지 않습니다. 세션 동안 파일이 진짜로 디스크에 존재하지 않습니다.

시크릿은 정상 작동합니다 — 캐시에서 주입과 마스킹에 사용됩니다. 세션 후 `blindenv cache-restore`로 복원할 수 있습니다.

---

## 설정

```yaml
# blindenv.yml

mode: stealth           # block (기본값) | stealth | evacuate

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
| `mode` | 보안 모드: `block`, `stealth`, `evacuate` | 기본 block 대신 stealth/evacuate를 원할 때 |
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
blindenv evacuate                     시크릿 파일을 디스크에서 삭제 (evacuate 모드 전용)
blindenv cache-restore                캐시에서 시크릿 파일 복원
blindenv cache-refresh                시크릿 파일 캐시 갱신 (.env 직접 수정 후)
blindenv hook cc <hook>               Claude Code PreToolUse 훅
                                       bash | read | grep | glob | guard-file
```

---

## 시크릿 매니저 그 너머

기존 시크릿 매니저는 **저장과 전달**을 해결합니다 — 시크릿이 어디에 있고, 프로세스에 어떻게 도달하는가. blindenv는 다른 문제를 풉니다: AI 에이전트가 프로세스를 실행할 때, **전달 이후에 무슨 일이 벌어지는가**.

| 기능 | 시크릿 매니저 | blindenv |
|---|---|---|
| 중앙 집중식 시크릿 저장 | 지원 | — (기존 `.env` 사용) |
| 런타임 프로세스 주입 | 지원 | 지원 |
| 출력 마스킹 | — | 지원 |
| 파일 은닉 (차단이 아님 — 보이지 않음) | — | 지원 |
| 내용 기반 차단 | — | 지원 |
| 설정 변조 방지 | — | 지원 |
| AI 에이전트 도구 훅 | — | 지원 |

시크릿 매니저와 blindenv는 상호 보완적입니다. 시크릿 매니저가 올바른 값을 `.env`나 CI 파이프라인에 넣어줍니다. blindenv는 명령을 실행하는 에이전트가 그 값을 **사용**하되 **보지 못하게** 합니다.

에이전트 인식 레이어가 없으면, 주입된 시크릿은 파일에서 읽히거나, stdout으로 출력되거나, 복사된 파일을 통해 유출될 수 있습니다 — 상위에서 아무리 잘 관리했더라도.

---

## 왜 .env를 그냥 쓰면 안 되나요?

AI 에이전트가 API 키가 필요합니다. 그래서 채팅에 붙여넣거나, `.env`에 넣고 에이전트가 읽게 합니다.

어느 쪽이든, 시크릿이 에이전트의 컨텍스트에 들어갑니다 — 프롬프트 인젝션 한 번이면 유출됩니다. 에이전트가 악의적일 필요도 없습니다. 속으면 되는 겁니다.

**"학습 Opt-out 했는데요?"** — 네. AI 프로바이더가 당신의 데이터를 학습에 쓰지 않을 수도 있고, 보존 기간이 지나면 약속대로 삭제할 수도 있습니다. 하지만 시크릿이 서버에 도달한 순간부터 삭제되는 순간까지, 그 사이에 무슨 일이 벌어질지는 아무도 모릅니다. 보안 사고, 내부 유출, 잘못된 백업 설정, 법적 요청. 그 타임라인은 당신이 통제할 수 없습니다.

그리고 다른 어떤 서비스보다도, 여기에 저장되는 것은 단순한 비밀번호가 아닙니다 — **완전한 맥락과 함께 저장된** 인증 정보입니다. 대화에는 그 키가 무엇을 위한 것인지, 어떤 서비스에 접근하는지, API를 어떻게 호출하는지, 어떤 인프라에 연결되는지가 모두 담겨 있습니다. 이 기록이 유출되는 순간, 단순한 키 유출이 아니라 완벽한 범행 도구가 됩니다.

유일한 방어는 애초에 시크릿을 보내지 않는 것입니다.

**"그냥 에이전트한테 API 키를 안 주면 되잖아요."** — 네. 발렛파킹에 차 키를 안 맡기면 되죠. 세 블록 떨어진 곳에 주차하고, 비 맞으며 짐을 들고, 보안 의식이 뛰어난 자신을 칭찬하면 됩니다. AI 에이전트의 존재 이유는 실제 작업을 하는 것입니다 — 코드 배포, API 호출, 서비스 접근. 인증 정보 없는 에이전트는 아주 비싼 자동완성입니다. 중요한 건 안 쓰는 게 아니라, 안전하게 쓰는 겁니다.

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
