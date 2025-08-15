# NonActiveX Server

Go로 작성된 파일 업로드/다운로드 웹 서버입니다.

## Docker 실행 방법

### 1. 빌드 및 실행
```bash
# Docker Compose로 빌드 및 실행
docker-compose up --build

# 백그라운드에서 실행
docker-compose up -d --build
```

### 2. 서비스 접속
- 웹 인터페이스: http://localhost:8443
- API 엔드포인트: http://localhost:8443/api/

### 3. 환경 변수 설정
다음 환경 변수들을 `docker-compose.yml`에서 수정할 수 있습니다:

- `SERVER_ADDR`: 서버 주소 (기본값: :8443)
- `UPLOAD_DIR`: 업로드 디렉토리 (기본값: ./uploads)
- `SERVER_API_TOKEN`: 서버 API 토큰 (기본값: change-me-server-token)
- `CLIENT_TOKEN`: 클라이언트 토큰 (기본값: change-me-client-token)

### 4. 볼륨 마운트
- `./uploads:/app/uploads`: 업로드된 파일들이 호스트의 `./uploads` 디렉토리에 저장됩니다.

### 5. 서비스 관리
```bash
# 서비스 중지
docker-compose down

# 로그 확인
docker-compose logs -f

# 서비스 재시작
docker-compose restart
```

## API 엔드포인트

- `GET /`: 웹 인터페이스
- `GET /api/files`: 업로드된 파일 목록 조회
- `GET /api/download/{filename}`: 파일 다운로드
- `POST /api/upload`: 파일 업로드 (X-Server-Token 헤더 필요)

## 보안

- 서버 API 토큰을 변경하여 업로드 엔드포인트를 보호합니다.
- 클라이언트 토큰을 변경하여 웹 인터페이스 보안을 강화합니다.
- 컨테이너는 non-root 사용자로 실행됩니다.
