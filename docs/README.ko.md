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
  <a href="../README.md">English</a> ·
  <strong>한국어</strong> ·
  <a href="./README.ja.md">日本語</a>
</p>

---

## 설치

먼저 마켓플레이스에서 플러그인을 추가합니다:

```
/plugin marketplace add neuradex/blindenv
```

그 다음 설치합니다 — 보호하고 싶은 폴더 안에서 설치하거나, 유저 스코프로 설치하면 모든 프로젝트에 적용됩니다:

```bash
# 프로젝트 스코프 — 이 프로젝트만 보호
cd /your/project
/plugin install blindenv@blindenv

# 유저 스코프 — 모든 프로젝트 보호
/plugin install blindenv@blindenv --user
```

Claude Code를 재시작하세요. 다음 세션부터 blindenv가 활성화됩니다.

---

## 동작 원리

에이전트가 `.env`를 읽으면 변수 이름과 주석은 보이지만, 모든 시크릿 값은 `[BLINDED]`로 가려집니다. 에이전트가 명령을 실행하면 blindenv가 실제 값을 서브프로세스에 보이지 않게 주입합니다. 시크릿이 포함된 출력도 자동으로 마스킹됩니다.

```
에이전트가 .env 읽기:  API_KEY=[BLINDED]
                      DB_URL=[BLINDED]
                      DEBUG=true        ← 시크릿 아닌 값은 그대로

에이전트가 명령 실행:  curl -H "Authorization: $API_KEY" https://api.example.com
                          ↓
blindenv 주입:         서브프로세스 env에 실제 API_KEY 값 주입
                          ↓
에이전트가 수신:       {"result": "ok", "token": "[BLINDED]"}
```

`secret_files`에 있는 값은 같은 이름의 셸 환경변수보다 우선합니다.

---

## blindenv.yml

첫 실행 시 프로젝트 루트에 `blindenv.yml`이 자동 생성됩니다. 생성되지 않은 경우 직접 만드세요:

```yaml
# blindenv.yml
secret_files:
  - .env
```

최소 구성은 이게 전부입니다. 더 추가하려면:

```yaml
# blindenv.yml
secret_files:         # 파싱, 마스킹, 서브프로세스 주입할 파일
  - .env
  - .env.local

mask_keys:            # 셸 환경변수를 이름으로 마스킹 (파일 없이 이미 export된 변수)
  - MY_CUSTOM_VAR

mask_patterns:        # 이름에 해당 문자열이 포함된 환경변수 마스킹
  - INTERNAL          # (기본값: KEY, SECRET, TOKEN, PASSWORD 등 포함)
```

**`secret_files` vs `mask_keys`:** 시크릿이 파일에 있으면 `secret_files`. 셸에 이미 export되어 있거나 CI/CD에서 주입된 변수라면 `mask_keys` — 파일 없이.

> 고급 옵션(`block` 모드, `passthrough` 등)은 [고급 설정](./ADVANCED.md)을 참고하세요.

---

## 설정 파일 탐색

blindenv는 현재 디렉토리에서 위로 올라가며 가장 가까운 `blindenv.yml`을 찾습니다 — `.gitignore`와 동일한 방식입니다. 상위 디렉토리의 설정은 하위 모든 프로젝트에 적용됩니다.

특정 폴더를 별도로 보호하고 싶다면, 그 폴더에 `blindenv.yml`을 두면 됩니다.

---

## 라이선스

MIT

---

<p align="center">
  <strong>에이전트는 API 키를 알 필요가 없습니다. 그냥 쓸 수 있으면 됩니다.</strong>
  <br>
  blindenv는 값을 숨긴 채로 작동하게 합니다.
</p>
