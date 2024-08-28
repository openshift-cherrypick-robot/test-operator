package horizontest

import (
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"

	testv1beta1 "github.com/openstack-k8s-operators/test-operator/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Job - prepare job to run Horizon tests
func Job(
	instance *testv1beta1.HorizonTest,
	labels map[string]string,
	jobName string,
	logsPVCName string,
	mountCerts bool,
	mountKeys bool,
	mountKubeconfig bool,
	envVars map[string]env.Setter,
) *batchv1.Job {

	runAsUser := int64(42455)
	runAsGroup := int64(42455)
	parallelism := int32(1)
	completions := int32(1)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism:  &parallelism,
			Completions:  &completions,
			BackoffLimit: instance.Spec.BackoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: instance.RbacResourceName(),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  &runAsUser,
						RunAsGroup: &runAsGroup,
						FSGroup:    &runAsGroup,
					},
					Containers: []corev1.Container{
						{
							Name:         instance.Name,
							Image:        instance.Spec.ContainerImage,
							Args:         []string{},
							Env:          env.MergeEnvs([]corev1.EnvVar{}, envVars),
							VolumeMounts: GetVolumeMounts(mountCerts, mountKeys, mountKubeconfig),
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{"NET_ADMIN", "NET_RAW", "CAP_AUDIT_WRITE"},
								},
								SeccompProfile: &corev1.SeccompProfile{
									Type: corev1.SeccompProfileTypeRuntimeDefault,
								},
							},
						},
					},
					Volumes: GetVolumes(
						instance,
						logsPVCName,
						mountCerts,
						mountKeys,
						mountKubeconfig,
					),
				},
			},
		},
	}

	return job
}
