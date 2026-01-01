package k8s_service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sService struct {
	dynamicClient dynamic.Interface
	mapper        meta.RESTMapper
}

var (
	instance *K8sService
	once     sync.Once
)

// params for deployment
type DeploymentParams struct {
	UserNamespace string
	NodePort      int32
}

// NewK8sService returns a singleton instance of K8sService
func NewK8sService() (*K8sService, error) {
	var err error
	once.Do(func() {
		// 1. Config 생성 (InCluster 우선, 실패시 로컬 kubeconfig)
		var config *rest.Config
		config, err = rest.InClusterConfig()
		if err != nil {
			// fallback to local kubeconfig
			kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
			if _, errExists := os.Stat(kubeconfig); errExists == nil {
				config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			}
		}
		if err != nil {
			err = fmt.Errorf("failed to get k8s config: %v", err)
			return
		}

		// 2. Dynamic Client 생성
		dynClient, errDyn := dynamic.NewForConfig(config)
		if errDyn != nil {
			err = fmt.Errorf("failed to create dynamic client: %v", errDyn)
			return
		}

		// 3. Discovery Client & Mapper 생성 (GVR 매핑용)
		dc, errDisc := discovery.NewDiscoveryClientForConfig(config)
		if errDisc != nil {
			err = fmt.Errorf("failed to create discovery client: %v", errDisc)
			return
		}
		mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

		instance = &K8sService{
			dynamicClient: dynClient,
			mapper:        mapper,
		}
	})

	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (s *K8sService) CheckConnectivity() (string, error) {
	// 간단한 연결 테스트 (System Namespaces 조회 시도)
	// GVR for Namespaces: v1, Namespace
	gvr, err := s.mapper.ResourceFor(schema.GroupVersionResource{Resource: "namespaces"})
	if err != nil {
		return "unhealthy", err
	}
	_, err = s.dynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{Limit: 1})
	if err != nil {
		return "unhealthy", err
	}
	return "healthy", nil
}

// checkInjection validates input parameters to prevent YAML injection and ensure K8s compatibility.
// 입력값 검증 함수: YAML 인젝션 방지 및 쿠버네티스 명명 규칙 준수 여부 확인
func (s *K8sService) checkInjection(userNamespace, vmName, password, dnsHost, manifestDir string, vmPort int32) error {
	// 1. 필수 파라미터 빈 값 체크
	// 모든 값이 설정되어 있어야 안전하게 템플릿 치환이 가능함
	if userNamespace == "" || vmName == "" || password == "" || dnsHost == "" || manifestDir == "" || vmPort == 0 {
		return fmt.Errorf("invalid parameters: empty values not allowed (필수 파라미터 누락)")
	}

	// 2. Port 범위 체크 (NodePort 범위: 30003-32767)
	// 사용자가 할당하려는 포트가 유효한 NodePort 범위 내에 있는지 확인
	if !(30003 <= vmPort && vmPort < 32767) {
		return fmt.Errorf("invalid port: %d (NodePort must be between 30003 and 32767)", vmPort)
	}

	// 3. DNS-1123 호환성 체크 (Namespace, VM Name)
	// 쿠버네티스 리소스 이름은 소문자, 숫자, '-' 만 허용하며, 숫자로 시작할 수 없음.
	// 정규식: ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// 여기서는 조금 더 느슨하게 검사하되, 특수문자나 공백, 뉴라인이 포함되지 않도록 함.
	dns1123Regex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

	if !dns1123Regex.MatchString(userNamespace) {
		return fmt.Errorf("invalid namespace format: %s (must be DNS-1123 compliant)", userNamespace)
	}
	if !dns1123Regex.MatchString(vmName) {
		return fmt.Errorf("invalid vmName format: %s (must be DNS-1123 compliant)", vmName)
	}

	// 4. Password 체크 (보안 및 인젝션 방지)
	// 길이: 최소 8자
	// 허용 문자: 알파벳 대소문자, 숫자, 특수문자 (!@#$%^&*()_+-=[]{}|;:,.<>/?)
	// 금지 문자: 공백, 뉴라인, 따옴표(" '), 역슬래시(\) 등 YAML 파싱에 문제될 수 있는 문자
	// 정규식: ^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{}|;:,.<>/?]{8,}$
	if len(password) < 8 || len(password) > 16 {
		return fmt.Errorf("invalid password: length must be between 8 and 16 characters")
	}

	// 안전한 문자셋 정의 (YAML Scalar로 안전하게 들어갈 수 있는 범위)
	passwordRegex := regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{}|;:,.<>/?]+$`)
	if !passwordRegex.MatchString(password) {
		return fmt.Errorf("invalid password format: contains invalid characters (allowed: alphanumeric and !@#$%%^&*()_+-=[]{}|;:,.<>/?)")
	}

	// 5. DNS Host 체크 (도메인 형식)
	// 알파벳, 숫자, '.', '-' 만 허용
	// 인젝션을 유발할 수 있는 ;, \n, && 등의 특수문자 차단
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	if !domainRegex.MatchString(dnsHost) {
		return fmt.Errorf("invalid dnsHost format: %s (contains invalid characters)", dnsHost)
	}

	// 6. Path Traversal 체크
	cleanDir := filepath.Clean(manifestDir)
	if strings.Contains(cleanDir, "..") || strings.HasPrefix(cleanDir, "/") || strings.HasPrefix(cleanDir, "\\") {
		// handle path traversal if needed
	}

	return nil
}

// CreatedResource holds metadata for tracking created objects
type CreatedResource struct {
	Group     string
	Version   string
	Kind      string
	Name      string
	Namespace string
	UID       types.UID
}

// VMInfo encapsulates the details of the created VM and resources
type VMInfo struct {
	Namespace        string
	Name             string
	Port             int32
	Password         string
	DNSHost          string
	CreatedResources []CreatedResource
}

// CreateUserVM creates resources defined in yaml-data/client-vm
func (s *K8sService) CreateUserVM(userNamespace, vmName, password, dnsHost, manifestDir string, vmPort int32) (*VMInfo, error) {
	// manifestDir := "yaml-data/client-vm" // 실행 위치 기준

	// Yaml에 그대로 넣지만, Injection검사를 시행.
	if err := s.checkInjection(userNamespace, vmName, password, dnsHost, manifestDir, vmPort); err != nil {
		return nil, err
	}

	//하드코딩된 yaml 파일들 폴더에 있는 것 모두 가져오기.
	files, err := os.ReadDir(manifestDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest directory: %v", err)
	}

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	vmInfo := &VMInfo{
		Namespace:        userNamespace,
		Name:             vmName,
		Port:             vmPort,
		Password:         password,
		DNSHost:          dnsHost,
		CreatedResources: []CreatedResource{},
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(manifestDir, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file.Name(), err)
		}

		// 텍스트 치환 (템플릿 처리)
		text := string(content)

		// 1. Namespace 치환 (YAML에 {{NAMESPACE}}가 있다면)
		text = strings.ReplaceAll(text, "{{NAMESPACE}}", userNamespace)

		// 2. NodePort 치환 (기존 30002 -> 입력받은 포트)
		text = strings.ReplaceAll(text, "{{NODEPORT}}", fmt.Sprintf("%d", vmPort))

		// 3. VM Name 치환 (기존 my-cloud-vps -> 입력받은 이름)
		text = strings.ReplaceAll(text, "{{VM_NAME}}", vmName)

		// 4. DNS Host 치환 (기존 vps.hy3on.site -> 입력받은 DNS)
		text = strings.ReplaceAll(text, "{{DNS_HOST}}", dnsHost)

		// 5. root 계정 Password 치환
		text = strings.ReplaceAll(text, "{{PASSWORD}}", password)

		// Multi-document support handling (e.g. separated by ---)
		docs := strings.Split(text, "\n---\n")
		for _, doc := range docs {
			if strings.TrimSpace(doc) == "" {
				continue
			}

			obj := &unstructured.Unstructured{}
			_, gvk, err := decUnstructured.Decode([]byte(doc), nil, obj)
			if err != nil {
				return nil, fmt.Errorf("failed to decode yaml in %s: %v", file.Name(), err)
			}

			// Namespace 강제 주입 (ClusterScoped 리소스 제외)
			// NetworkPolicy, Service, VirtualMachine, DataVolume 등은 Namespaced임.
			// ResourceQuota도 Namespaced.
			// YAML에 namespace 필드가 없거나, 있더라도 덮어씀.
			obj.SetNamespace(userNamespace)

			// Mapping GVK to GVR
			mapping, err := s.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
			if err != nil {
				return nil, fmt.Errorf("failed to find mapping for %s: %v", gvk.String(), err)
			}

			var dri dynamic.ResourceInterface
			if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
				dri = s.dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
			} else {
				dri = s.dynamicClient.Resource(mapping.Resource)
			}

			// Apply (Server-Side Apply 추천하지만, 간단히 Create or Update 전략 사용)
			// 여기서는 단순 Create 시도 후 이미 존재하면 Update 시도 로직 등을 쓸 수 있으나,
			// "VM제공하는 클라우드 서비스" 특성상 생성 실패시 에러가 나을 수 있음.
			// 다만, 멱등성을 위해 Apply(Patch)를 사용하는 것이 좋음.

			// data := []byte(doc) // Original json/yaml data needed for patch? No, unstructured obj is enough
			// Apply using ServerSideApply
			// force := true
			// _, err = dri.Patch(context.Background(), obj.GetName(), types.ApplyPatchType, []byte(doc), metav1.PatchOptions{FieldManager: "vm-controller", Force: &force})

			// 간단하게 Create 먼저 시도.
			createdObj, err := dri.Create(context.Background(), obj, metav1.CreateOptions{})
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					// 이미 존재하면 넘어가거나, Update 로직?
					// VM같은 경우 상태가 있어서 함부로 Update하면 리스타트 될 수 있음.
					// 우선 Log만 찍고 스킵.
					fmt.Printf("Resource %s %s already exists, skipping.\n", gvk.Kind, obj.GetName())
					// 이미 존재하더라도 Get을 통해 정보를 가져오는 것이 추적에 도움이 될 수 있음.
					// 여기서는 간단히 Skip. 필요하면 Get 호출 추가 가능.
				} else {
					return nil, fmt.Errorf("failed to create resource %s: %v", gvk.Kind, err)
				}
			} else {
				fmt.Printf("Successfully created %s: %s\n", gvk.Kind, obj.GetName())
				// 생성된 리소스 정보 저장
				vmInfo.CreatedResources = append(vmInfo.CreatedResources, CreatedResource{
					Group:     gvk.Group,
					Version:   gvk.Version,
					Kind:      gvk.Kind,
					Name:      createdObj.GetName(),
					Namespace: createdObj.GetNamespace(),
					UID:       createdObj.GetUID(),
				})
			}
		}
	}

	return vmInfo, nil
}
