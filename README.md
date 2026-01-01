# VM Controller (Cloud Backend)

## 📖 프로젝트 소개 (Project Overview)
**VM Controller**는 사용자가 온디맨드(On-Demand)로 가상 머신(VM) 및 컨테이너 기반 웹 서비스를 프로비저닝하고 관리할 수 있도록 지원하는 **클라우드 인프라 오케스트레이션 백엔드**입니다.

Kubernetes와 KubeVirt 기술을 기반으로 하여, 클라우드 네이티브 환경에서 VM과 컨테이너 워크로드를 통합 제어하며, 사용자에게 직관적인 API를 제공하여 인프라 복잡성을 추상화합니다.

---

## 🏗️ 아키텍처 및 기술 스택 (Architecture & Tech Stack)
본 프로젝트는 **MSA(Microservices Architecture)** 지향적인 설계를 따르며, 안정성과 확장성을 고려하여 다음과 같은 기술을 채택하였습니다.

### 1. Backend Core
*   **Language**: **Go (Golang) 1.20+** - 고성능 동시성 처리에 최적화된 언어로, 대규모 트래픽과 시스템 리소스 제어에 탁월합니다.
*   **Framework**: **Gin Web Framework** - 경량화된 고성능 HTTP 웹 프레임워크로, 낮은 지연 시간(Latency)을 보장합니다.
*   **Database & ORM**: **PostgreSQL** / **GORM** - 데이터 무결성을 보장하는 관계형 데이터베이스와 객체 매핑을 위한 ORM을 사용합니다.
*   **Authentication**: **JWT (JSON Web Token)** - Stateless한 인증 방식을 통해 확장성 있는 보안 체계를 구축했습니다.

### 2. Infrastructure & Virtualization
*   **Orchestration**: **Kubernetes (K8s) Distro (K3s)** - 컨테이너 오케스트레이션의 표준으로, 서비스의 자동 배포 및 스케일링을 담당합니다.
*   **Virtualization**: **KubeVirt** - Kubernetes 환경 네이티브하게 가상 머신(VM)을 실행하고 관리할 수 있게 해주는 확장(Extension)입니다.
*   **Storage**: **CDI (Containerized Data Importer)** - PVC(Persistent Volume Claim) 기반의 디스크 이미지 프로비저닝을 담당합니다.

---

## 🔑 주요 기능 (Key Features)
*   **VM Lifecycle Management**: 가상 머신의 생성, 조회, 중지, 삭제 및 리소스 크기 조절(Resize).
*   **Web Service Deployment**: 사용자 소스 코드 기반의 웹 애플리케이션 자동 배포 파이프라인.
*   **Resource Isolation**: Namespace 기반의 멀티 테넌시(Multi-tenancy) 환경 제공.
*   **Security**: 비밀번호 해싱(Bcrypt) 및 JWT 기반의 안전한 API 접근 제어.

---

## 🚀 향후 로드맵 (Roadmap)
현재 MVP(Minimum Viable Product) 단계이며, 엔터프라이즈급 기능을 목표로 고도화 예정입니다.

- [ ] **Advanced Networking**: Service Mesh 도입 및 정교한 Ingress/Egress 트래픽 제어.
- [ ] **Observability**: Prometheus & Grafana 연동을 통한 실시간 리소스 모니터링 대시보드 구축.
- [ ] **CI/CD Integration**: Tekton 또는 ArgoCD를 활용한 GitOps 기반 배포 파이프라인 구축.
- [ ] **RBAC (Role-Based Access Control)**: 세분화된 권한 관리를 통해 조직 내 보안 정책 강화.

---

## �️ 시작하기 (Getting Started)

### 사전 요구사항 (Prerequisites)
*   Go 1.20 이상
*   Kubernetes Cluster (K3s 권장)
*   KubeVirt 및 CDI 설치 완료

### 설치 및 실행
1.  **레포지토리 클론**
    ```bash
    git clone https://github.com/Hy3ons/cloud-backend.git
    cd cloud-backend
    ```

2.  **의존성 설치**
    ```bash
    go mod tidy
    ```

3.  **환경 설정**
    ```bash
    cp .env.example .env
    # .env 파일 내용을 환경에 맞게 수정
    ```

4.  **서버 실행**
    ```bash
    go run cmd/server/main.go
    ```
