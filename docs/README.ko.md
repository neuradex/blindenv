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

에이전트가 `.env` 파일을 읽으면 구조, 변수 이름, 주석은 보이지만 — 모든 시크릿 값은 `[BLINDED]`로 치환되어 있습니다. 명령은 실제 인증 정보가 뒤에서 주입된 채 실행되고, 시크릿이 포함된 출력은 자동으로 마스킹됩니다.

```
에이전트가 .env를 읽으면:  API_KEY=[BLINDED]
                          DB_PASSWORD=[BLINDED]
                          DEBUG=true              ← 시크릿이 아닌 값은 그대로

에이전트가 명령 실행:     curl -H "Authorization: $API_KEY" https://api.example.com
                              ↓
blindenv 프록시:          서브프로세스 환경에 실제 값 주입
                              ↓
에이전트가 수신:          {"result": "ok", "token": "[BLINDED]"}
```

### 에이전트가 보는 것

| 에이전트 동작 | 결과 |
|---|---|
| `.env` 읽기 | 변수 이름은 보임, 모든 값 → `[BLINDED]` |
| Bash에서 `echo $API_KEY` | `[BLINDED]` |
| `$API_KEY`로 `curl` 실행 | 정상 동작 — 서브프로세스에 실제 값 주입 |
| 시크릿이 포함된 출력 | 자동으로 `[BLINDED]`로 치환 |

에이전트는 프로젝트 설정 구조에 대한 완전한 맥락을 얻습니다. 다만 중요한 값은 볼 수 없을 뿐입니다. 그러면서도 API 호출은 성공하고, 배포는 완료되고, 서비스는 응답합니다.

---

## 설치

### Claude Code 플러그인 (권장)

```bash
/plugin marketplace add neuradex/blindenv
/plugin install blindenv@blindenv
```

끝입니다. 다음 세션 시작 시 플랫폼에 맞는 바이너리가 [GitHub Releases](https://github.com/neuradex/blindenv/releases)에서 자동 다운로드되고, 프로젝트 루트에 `blindenv.yml`이 자동 생성됩니다.

`blindenv.yml`을 열어 시크릿이 포함된 파일을 설정하세요:

```yaml
secret_files:
  - .env
  - .env.local
  # - ~/.aws/credentials
```

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

`blindenv.yml`이 있으면 모든 것이 자동으로 작동합니다:

| | 동작 |
|---|---|
| **마스킹** | 에이전트가 `.env` 파일을 읽으면 모든 값이 `[BLINDED]`로 치환 |
| **주입** | `blindenv run`을 통해 명령에서 `$VAR`로 실제 시크릿 값 사용 가능 |
| **출력 마스킹** | 시크릿 값이 포함된 출력 → `[BLINDED]` |

```bash
# .env 내용: API_KEY=sk-a1b2c3d4
blindenv run 'curl -H "Authorization: Bearer $API_KEY" https://api.example.com'
# → {"result": "ok", "key": "[BLINDED]"}
```

Claude Code 플러그인으로 사용 시에는 `blindenv run`을 직접 쓸 필요도 없습니다 — 훅이 Bash 명령을 자동으로 래핑합니다.

---

## 동작 원리

### 방어 계층

| # | 계층 | 역할 |
|---|------|------|
| 1 | **파일 마스킹** | 시크릿 파일은 읽을 수 있지만, 모든 값이 `[BLINDED]`로 치환 |
| 2 | **서브프로세스 격리** | 실제 시크릿은 서브프로세스 환경에만 존재 — 에이전트 컨텍스트에는 절대 노출되지 않음 |
| 3 | **출력 마스킹** | stdout/stderr를 스캔하여 시크릿 값을 `[BLINDED]`로 치환 |
| 4 | **자동 감지** | `KEY`, `SECRET`, `TOKEN` 등의 패턴에 매칭되는 환경변수를 자동으로 마스킹 |
| 5 | **설정 보호** | 에이전트가 `blindenv.yml`을 수정할 수 없음 — 변조 불가 |

### Claude Code 훅

플러그인 설치 시, 6개의 PreToolUse 훅이 모든 에이전트 동작을 감시합니다:

| 도구 | 훅 | 동작 |
|------|-----|------|
| **Bash** | `blindenv hook cc bash` | 명령을 `blindenv run '...'`으로 래핑 — 시크릿 주입, 출력 마스킹 |
| **Read** | `blindenv hook cc read` | 시크릿 파일 → 모든 값이 `[BLINDED]`로 마스킹된 사본 |
| **Edit** | `blindenv hook cc guard-file` | 시크릿 파일과 `blindenv.yml` → 수정 보호 |
| **Write** | `blindenv hook cc guard-file` | 시크릿 파일과 `blindenv.yml` → 수정 보호 |
| **Grep** | `blindenv hook cc grep` | 시크릿 파일이 검색 결과에서 제외 |
| **Glob** | `blindenv hook cc glob` | 시크릿 파일이 파일 목록에서 제외 |

---

## 설정

```yaml
# blindenv.yml

secret_files:            # .env 파일 — 자동 파싱, 값 마스킹
  - .env
  - .env.local
  - ~/.aws/credentials

# mask_patterns:         # 자동 감지를 위한 환경변수 이름 패턴
#   - KEY               # (생략 시 기본값 적용 — KEY, SECRET, TOKEN 등)
#   - SECRET

# mask_env:              # 명시적으로 마스킹할 환경변수 (패턴에 매칭되지 않는 이름용)
#   - MY_CUSTOM_VAR

# inject:                # 호스트 프로세스의 환경변수 — 주입 + 마스킹
#   - CI_TOKEN
#   - DEPLOY_KEY
```

| 필드 | 용도 | 사용 시점 |
|------|------|----------|
| `secret_files` | `.env` 파일 파싱, 값 마스킹, 파일 접근 차단 | **항상** — 핵심 메커니즘 |
| `mask_patterns` | 이름 패턴으로 환경변수 자동 감지 (예: `KEY`, `TOKEN`) | 생략 시 기본값 적용; 범위를 좁히거나 넓힐 때 커스터마이즈 |
| `mask_env` | 명시적으로 마스킹할 환경변수 이름 | 패턴에 매칭되지 않는 변수 |
| `inject` | 호스트 프로세스에서 환경변수를 서브프로세스에 주입 | CI/CD 시크릿, 파일에 없는 변수 |

**`secret_files`만으로도 충분합니다.** 기본 `mask_patterns`가 프로세스 환경에서 일반적인 시크릿 변수 이름(`KEY`, `SECRET`, `TOKEN`, `PASSWORD` 등)을 자동으로 감지합니다.

설정 파일은 `cwd`에서 `/`까지 상위로 탐색한 뒤 `~/.blindenv.yml`도 확인합니다. 가장 가까운 파일이 적용됩니다 — `.gitignore`와 같은 방식입니다.

> 추가 보안 모드(`block`, `stash`)와 고급 옵션(`passthrough`)은 [고급 설정](./ADVANCED.md)을 참고하세요.

---

## CLI 레퍼런스

```
blindenv run '<command>'              시크릿 격리 + 출력 마스킹으로 실행
blindenv init                         현재 디렉토리에 blindenv.yml 생성
blindenv hook cc <hook>               Claude Code PreToolUse 훅
                                       bash | read | grep | glob | guard-file
```

> 추가 명령어(`stash`, `cache-restore`, `cache-refresh`, `check-file`)는 [고급 설정](./ADVANCED.md)을 참고하세요.

---

## 시크릿 매니저 그 너머

기존 시크릿 매니저는 **저장과 전달**을 해결합니다 — 시크릿이 어디에 있고, 프로세스에 어떻게 도달하는가. blindenv는 다른 문제를 풉니다: AI 에이전트가 프로세스를 실행할 때, **전달 이후에 무슨 일이 벌어지는가**.

| 기능 | 시크릿 매니저 | blindenv |
|---|---|---|
| 중앙 집중식 시크릿 저장 | 지원 | — (기존 `.env` 사용) |
| 런타임 프로세스 주입 | 지원 | 지원 |
| 출력 마스킹 | — | 지원 |
| 파일 마스킹 (`[BLINDED]` 값) | — | 지원 |
| 변수 이름으로 자동 감지 | — | 지원 |
| 설정 변조 방지 | — | 지원 |
| AI 에이전트 도구 훅 | — | 지원 |

시크릿 매니저와 blindenv는 상호 보완적입니다. 시크릿 매니저가 올바른 값을 `.env`나 CI 파이프라인에 넣어줍니다. blindenv는 명령을 실행하는 에이전트가 그 값을 **사용**하되 **보지 못하게** 합니다.

---

## 왜 .env를 그냥 쓰면 안 되나요?

AI 에이전트가 API 키가 필요합니다. 그래서 채팅에 붙여넣거나, `.env`에 넣고 에이전트가 읽게 합니다.

어느 쪽이든, 시크릿이 에이전트의 컨텍스트에 들어갑니다 — 프롬프트 인젝션 한 번이면 유출됩니다. 에이전트가 악의적일 필요도 없습니다. 속으면 되는 겁니다.

**"학습 Opt-out 했는데요."** — 네. AI 프로바이더가 당신의 데이터를 학습에 쓰지 않을 수도 있고, 보존 기간이 지나면 약속대로 삭제할 수도 있습니다. 하지만 시크릿이 서버에 도달한 순간부터 삭제되는 순간까지, 그 사이에 무슨 일이 벌어질지는 아무도 모릅니다. 보안 사고, 내부 유출, 잘못된 백업 설정, 법적 요청. 그 타임라인은 당신이 통제할 수 없습니다.

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
