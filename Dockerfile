# ------------------------------------------------
# builder image
# ------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

# 아키텍처 ARG 선언 (빌드 시 전달받음)
ARG TARGETARCH=amd64
ARG BINARY_NAME=service

# 아키텍처 정보 출력
RUN echo "빌드 정보:" && \
    echo "  Target Architecture: ${TARGETARCH}" && \
    echo "  Binary Name: ${BINARY_NAME}" && \
    echo "  선택될 바이너리: ${BINARY_NAME}-linux-${TARGETARCH}"

# 필수 도구 설치
RUN apk add --no-cache upx ca-certificates tzdata putty-tools

WORKDIR /app

COPY ${BINARY_NAME}-linux-${TARGETARCH} ./service
COPY public/ ./public/

# Go 모듈 파일 먼저 복사
#COPY go.mod go.sum ./

# Go 모듈 다운로드
#RUN go mod download

# 소스 코드 전체 복사
#COPY . .

# 보안 강화 빌드 (cmd 디렉토리에서 빌드, 빌드 된걸 전달 받음)
#RUN CGO_ENABLED=0 GOOS=linux go build \
#    -ldflags="-s -w -extldflags '-static'" \
#    -a -installsuffix cgo \
#    -o service ./cmd

# 보안 강화 빌드 (cmd 디렉토리에서 빌드)
# 빌드 시점에 버전 정보 주입
#RUN CGO_ENABLED=0 GOOS=linux go build \
#    -ldflags="-s -w -extldflags '-static' \
#              -X main.version=${VERSION} \
#              -X main.buildTime=${BUILD_DATE} \
#              -X main.gitCommit=${GIT_COMMIT} \
#              -X main.gitBranch=${GIT_BRANCH}" \
#    -a -installsuffix cgo \
#    -o service ./cmd

# 바이너리 정보 확인
#RUN ls -la service && \
#    file service && \
#    echo "✅ 바이너리 빌드 완료: ${BINARY_NAME}"

# 바이너리 압축/난독화 : 시간 걸려서 주석처리
# RUN upx --best --ultra-brute service

# 실행 권한 설정
RUN chmod +x service


# ------------------------------------------------
# runtime image
# ------------------------------------------------
FROM scratch

# 🏷️ 런타임 빌드 인수
ARG TARGETARCH=amd64
ARG BINARY_NAME=service
ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown
ARG GIT_BRANCH=unknown

# 필수 파일 복사
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /usr/bin/puttygen /usr/bin/

# 바이너리 파일 복사
COPY --from=builder /app/service /service

# public 디렉토리 복사 (정적 파일들)
COPY --from=builder /app/public /public

# 메타데이터
LABEL maintainer="Control System Team"
LABEL security.distroless="true"
LABEL security.tools="puttygen"
LABEL app.binary="${BINARY_NAME}"

# 포트 노출
EXPOSE 8080

# 고정된 서비스 실행
ENTRYPOINT ["/service"]