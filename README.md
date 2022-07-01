# podprinter
 - CRD 작업 사항 template 코드 떨구기
    ```sh
    make manifests
    ```
 
 - CRD 를 ETCD로 밀어넣기
    ```sh
    make install
    ```
 
 - Custom Controller 실행 (main.go 파일)
    ```
    make run
    ```
 
 - CR 배포
    ```sh
    kubectl apply -f config/samples/mygroup_v1_hello.yaml
    ```
 
## Description
 - CR 배포 네임스페이스 => 현재 kubectl 이 바라보고 있는 namespace context
 - Custom Controller
   - CR Watch 범위: 모든 네임스페이스 (별도 제한 범위 설정 X)
   - CR Event 감지시 Logs 출력
   - CR 의 name 과 namespace 를 토대로 Deployment 존재 유무 검사
     - 없을 경우 Deployment 배포, 배포된 pod는 CR의 spec.msg 출력
   - CR의 spec.size 와 deployment의 replicaSet 개수 일치 여부 검사
     - spec.size 만큼 deployment의 replicaSet 개수 동기화
