package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func CreateChildProcessHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}

		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "child-process-",
			},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyNever,
				Containers: []v1.Container{
					{
						Name:  "child-process",
						Image: "lak9348/child-process:v1.0.2",
					},
				},
			},
		}

		// Create Pod
		fmt.Println("Creating Pod...")
		result, err := clientSet.CoreV1().Pods("default").Create(
			context.Background(),
			pod,
			metav1.CreateOptions{},
		)
		fmt.Printf("Created Pod %q.\n", result.GetObjectMeta().GetName())

		ctx.JSON(http.StatusOK, "Pod created successfully!")
	}
}
