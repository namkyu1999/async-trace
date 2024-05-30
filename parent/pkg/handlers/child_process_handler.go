package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// ref: https://github.com/tektoncd/pipeline/blob/1f3b1b485b25d91e2b04b3f35a1a7ff2494bacd1/pkg/reconciler/pipelinerun/tracing.go#L84
// Extract spanContext from the context and return it as json encoded string
func getMarshalledSpanFromContext(ctx context.Context) (string, error) {
	carrier := make(map[string]string)
	pro := otel.GetTextMapPropagator()

	pro.Inject(ctx, propagation.MapCarrier(carrier))

	if len(carrier) == 0 {
		return "", errors.New("spanContext not present in the context, unable to marshall")
	}

	marshalled, err := json.Marshal(carrier)
	if err != nil {
		return "", err
	}
	if len(marshalled) >= 1024 {
		return "", errors.New("marshalled spanContext size is too big")
	}
	return string(marshalled), nil
}

// For Docker
func CreateChildProcessHandler(tracer trace.Tracer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		spanContext, span := tracer.Start(ctx, "create-child-process")
		defer span.End()

		traceParent, err := getMarshalledSpanFromContext(spanContext)
		if err != nil {
			panic(err)
		}

		cli, err := client.NewClientWithOpts()
		if err != nil {
			panic(err)
		}
		defer cli.Close()
		cli.NegotiateAPIVersion(spanContext)

		imageName := "lak9348/child-process:v1.0.15"

		out, err := cli.ImagePull(spanContext, imageName, image.PullOptions{})
		if err != nil {
			panic(err)
		}
		defer out.Close()
		io.Copy(os.Stdout, out)

		resp, err := cli.ContainerCreate(spanContext, &container.Config{
			Image:    imageName,
			Hostname: "childservice",
			Env: []string{
				"OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317",
				"TRACE_PARENT=" + traceParent,
			},
		}, &container.HostConfig{
			NetworkMode: "async-trace",
		}, nil, nil, "")
		if err != nil {
			panic(err)
		}

		if err := cli.ContainerStart(spanContext, resp.ID, container.StartOptions{}); err != nil {
			panic(err)
		}
		log.Info("Container started successfully!")

		ctx.JSON(200, "Container started successfully!")
	}
}

// For Kubernetes
//func CreateChildProcessHandler() gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		config, err := rest.InClusterConfig()
//		if err != nil {
//			panic(err)
//		}
//
//		clientSet, err := kubernetes.NewForConfig(config)
//		if err != nil {
//			panic(err)
//		}
//
//		pod := &v1.Pod{
//			ObjectMeta: metav1.ObjectMeta{
//				GenerateName: "child-process-",
//			},
//			Spec: v1.PodSpec{
//				RestartPolicy: v1.RestartPolicyNever,
//				Containers: []v1.Container{
//					{
//						Name:  "child-process",
//						Image: "lak9348/child-process:v1.0.2",
//					},
//				},
//			},
//		}
//
//		// Create Pod
//		fmt.Println("Creating Pod...")
//		result, err := clientSet.CoreV1().Pods("default").Create(
//			context.Background(),
//			pod,
//			metav1.CreateOptions{},
//		)
//		fmt.Printf("Created Pod %q.\n", result.GetObjectMeta().GetName())
//
//		ctx.JSON(http.StatusOK, "Pod created successfully!")
//	}
//}
