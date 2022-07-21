package controllers

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mygroupv1 "github.com/hojun121/podprinter-operator/api/v1"
)

// HelloReconciler reconciles a Hello object
type HelloReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *HelloReconciler) Reconcile(ctx context.Context, req ctrl.Request) (emptyResult ctrl.Result, err error) {

	// cr로 정의한 객체를 가져오기위한 struct의 ref
	myCustomResource := &mygroupv1.Hello{}

	// 데이터를 서버에서 받아와 myCustomResource에 Input
	if err = r.Client.Get(ctx, req.NamespacedName, myCustomResource); err != nil {
		// 변경사항인 cr이 k8s에 존재하는지를 확인
		if errors.IsNotFound(err) {
			// cr이 삭제됨
			return emptyResult, nil
		}
		// GET함수 에러처리
		return emptyResult, err
	}

	// Print Spec.Msg Of Custom Resource (CR)
	fmt.Println(fmt.Sprintf("Here is Operator Log: %s", myCustomResource.Spec.Msg))

	// CR의 정보 기반 디플로이먼트 객체 유무 검사
	myDeployment := &appsv1.Deployment{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      myCustomResource.Name,
		Namespace: myCustomResource.Namespace,
	}, myDeployment); err != nil {
		// 없다면
		if errors.IsNotFound(err) {
			if err = r.Create(ctx, r.createDeployment(myCustomResource)); err != nil {
				return emptyResult, err
			}
			return ctrl.Result{RequeueAfter: time.Second * 2}, nil
			// 이벤트큐에 다시 올라가 로직 재실행 방법 => (1) ctrl.Result의 Requeue를 true로 설정 (2) RequeueAfter 시간 지정
		}
		return emptyResult, err
	}

	// Deployment를 생성할때 cr의 size로 replicaset 생성
	// CR의 Spec.size 값이 변경되면 감지하여 Deployment 업데이트
	mySize := myCustomResource.Spec.Size
	// if count(deployment.replicaset) != cr.size
	if *myDeployment.Spec.Replicas != mySize {
		myDeployment.Spec.Replicas = &mySize
		// custom controller를 만들더라도 기존 k8s control loop는 정상 동작. replicaset만 변경해서 pod 수를 제어
		err = r.Client.Update(ctx, myDeployment)
		if err != nil {
			return emptyResult, err
		}
		return ctrl.Result{RequeueAfter: time.Second * 2}, nil
	}
	return emptyResult, nil
}

// Deployment를 생성하고 컨틀롤러에 등록해 cr이 삭제된경우 함께 삭제
func (r *HelloReconciler) createDeployment(m *mygroupv1.Hello) *appsv1.Deployment {
	myLabel := getLabelForMyCustomResource(m.Name)

	mySize := m.Spec.Size

	newDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &mySize,
			Selector: &metav1.LabelSelector{
				MatchLabels: myLabel,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: myLabel,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "busybox",
						Name:    m.Name,
						Command: []string{"/bin/echo", m.Spec.Msg},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(m, newDeployment, r.Scheme)
	return newDeployment
}

func getLabelForMyCustomResource(name string) map[string]string {
	return map[string]string{"app": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// For에 감시할 cr을 설정해주고
	// Owns는 서브로 감시할 대상을 설정합니다.(서브로 감시하는 대상이 삭제된경우 reconcile되도록)
	// 서브로 감시할 대상에 추가된 service와 deployment는
	// 추후 사용자가 임의로 삭제했을때 다시 복구됩니다.
	// cr이 삭제됐을때 svc와 dep가 함께 삭제 => 컨트롤러에 ref 추가
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1.Hello{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
