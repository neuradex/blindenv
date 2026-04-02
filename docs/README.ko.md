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

```bash
/plugin marketplace add neuradex/blindenv
/plugin install blindenv@blindenv
/reload-plugins
```

끝입니다. 이 순간부터 에이전트는 `.env` 파일의 전체 구조를 읽을 수 있지만, 모든 시크릿 값은 `[BLINDED]`로 가려집니다. 마법 같은 점: Claude Code는 명령을 실행할 때 그 값들을 여전히 사용할 수 있습니다. 실제 값은 서브프로세스에 보이지 않게 주입되고, 시크릿이 포함된 출력도 자동으로 마스킹됩니다.

---

## 더 세밀하게 제어하고 싶다면?

프로젝트 루트에 `blindenv.yml`이 자동 생성됩니다. 열어서 보호할 파일을 설정하세요:

```yaml
secret_files:
  - .env
  - .env.local
  - secrets.yaml
```

특정 환경변수를 명시적으로 마스킹하거나 이름 패턴을 추가할 수도 있습니다:

```yaml
secret_files:
  - .env

mask_env:
  - MY_CUSTOM_VAR      # 특정 변수 마스킹

mask_patterns:
  - KEY                # 이름에 "KEY"가 포함된 모든 환경변수 마스킹
  - TOKEN
```

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
  <strong>에이전트에게 필요한 건 당신의 키가 아닙니다. 키가 여는 것입니다.</strong>
  <br>
  blindenv는 노출 없이 접근을 제공합니다.
</p>
