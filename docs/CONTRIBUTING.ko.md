<!-- AGENT NOTE: 이 파일을 수정하면 CONTRIBUTING.md (영어)와 docs/CONTRIBUTING.ja.md (일본어)도 함께 업데이트하세요. -->

# blindenv에 기여하기

<p align="center">
  <a href="../CONTRIBUTING.md">English</a> ·
  <strong>한국어</strong> ·
  <a href="./CONTRIBUTING.ja.md">日本語</a>
</p>

## 사전 요구사항

- Go 1.22+
- Make

## 개발 환경 구성

```bash
git clone https://github.com/neuradex/blindenv.git
cd blindenv
make build
```

프로젝트 루트에 `./blindenv` 바이너리가 생성됩니다.

## 주요 명령어

```bash
make build    # 로컬 바이너리 빌드 (git 태그에서 버전 주입)
make test     # 전체 테스트 실행
make vet      # go vet 실행
make clean    # 빌드된 바이너리 삭제
make purge    # 시스템에서 blindenv 흔적 전체 삭제 (설치 테스트용)
```

## 버전 관리 & 릴리스

버전은 **두 곳**에 존재하며, 관리 방식이 다릅니다:

| 위치 | 결정 주체 | 업데이트 시점 |
|------|----------|-------------|
| Go 바이너리 (`blindenv version`) | Git 태그 (빌드 시 `-ldflags`로 주입) | 빌드 시 자동 |
| `plugin.json` / `marketplace.json` | `make bump` | 태그 전에 수동 |

**Go 바이너리에 버전이 하드코딩되어 있지 않습니다.** 빌드 시 git 태그에서 자동 주입됩니다 — 로컬(`make build`)과 CI(GoReleaser) 모두. 태그 직후 `make build`하면 `v0.4.0` 같은 깔끔한 버전이 나오고, 태그 이후 커밋이 있으면 `v0.4.0-3-gabcdef`처럼 표시됩니다.

### 릴리스 플로우

```bash
make bump v=0.4.0                      # plugin.json + marketplace.json 업데이트
git add -A && git commit -m "chore: v0.4.0"
git tag v0.4.0                          # 이 태그가 바이너리 버전을 결정
git push origin main --tags             # GoReleaser 트리거 → GitHub Release
```

`make bump`을 실행하면 복사해서 쓸 수 있는 git 명령어가 출력됩니다.

## 아키텍처

blindenv는 두 가지 실행 모드를 가집니다:

1. **`blindenv run '<cmd>'`** — 시크릿이 주입되고 출력이 마스킹되는 격리된 서브프로세스에서 명령을 실행합니다.
2. **`blindenv hook cc <hook>`** — Claude Code 도구 호출을 실행 전에 가로채는 PreToolUse 훅 핸들러입니다.

훅은 stdin에서 JSON을 읽고 (Claude Code 훅 프로토콜), 보안 로직을 적용한 뒤, stdout/stderr + 종료 코드를 통해 allow/block/modify 액션으로 응답합니다.

`provider` 패키지가 훅 프로토콜을 추상화하고 있어, 다른 AI 코딩 에이전트(예: Cursor, Windsurf)를 지원하려면 `Provider` 인터페이스만 구현하면 됩니다.

## 새 Provider 추가하기

1. `provider/<name>/<name>.go`에 `provider.Provider` 인터페이스 구현
2. `cmd/hook.go`의 `resolveProvider()`에 등록
3. 해당 플랫폼의 훅 설정 추가

## 테스트

```bash
make test
```

테스트는 소스 파일과 같은 위치에 있습니다 (`engine/*_test.go`). 새 기능을 추가할 때는 최소한 engine 레이어의 테스트를 포함해주세요.

## 변경사항 제출

1. 레포를 포크하고 기능 브랜치를 생성
2. 변경 작업 수행
3. `make test && make vet` 실행
4. 무엇을 왜 바꿨는지 명확한 설명과 함께 PR 생성
