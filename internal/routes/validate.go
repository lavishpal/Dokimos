package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	dockerHubAllowedNamespaces = []string{"default", "kube-system", "dev"} // Example allowed namespaces
	admissionReview            admissionv1.AdmissionReview                 // This is the struct that represents the AdmissionReview object
	podObject                  corev1.Pod                                  // This is the struct that represents the Pod object in the AdmissionReview object
	admissionResponse          admissionv1.AdmissionResponse               // This is the struct that represents the AdmissionResponse object
	isAllowed                  bool
	// Result is a pointer to metav1.Status so that we can modify its fields (like Code and Message)
	// and have those changes reflected wherever Result is used, without copying the struct.
	Result = &metav1.Status{}
)

func AdmissionReviewValidate(c *gin.Context) {
	isAllowed = true
	Result = &metav1.Status{
		Code:    200, // HTTP status code for success
		Message: fmt.Sprintf("Request allowed for Namespace: %s", podObject.Namespace),
	}

	body, err := c.GetRawData()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read request body"})
		return
	}
	// Body is the raw request body of the api call to our webhook, which is expected to be an AdmissionReview object
	// Unmarshel means to convert the JSON body into an AdmissionReview struct so that we can work with it in Go.

	if err := json.Unmarshal(body, &admissionReview); err != nil {
		log.Printf("Error unmarshaling AdmissionReview: %v", err)
		return
	}

	// log.Printf("Successfully unmarshaled AdmissionReview for UID: %s", admissionReview.Request.UID)

	// log.Printf("Request Kind: %s, Namespace: %s, Name: %s",
	// 	admissionReview.Request.Kind.Kind,
	// 	admissionReview.Request.Namespace,
	// 	admissionReview.Request.Name,
	// )
	// The Object's ( pod, configmap, etc. ) raw body sent in the request is stored in admissionReview.Request.Object.Raw.
	// It is a byte slice ( []byte ) that contains the JSON representation of the object.
	// We need to Unmarshel ( store the raw json string into a datastrucutre, as we did before )
	// the raw data of admissionReview.Request.Object.Raw into a Pod object, because we know it is a pod,

	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &podObject); err != nil {
		log.Printf("Error unmarshaling Pod object: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to unmarshal pod object"})
		return
	}

	// log.Printf("Successfully unmarshaled Pod object for Name: %s, Namespace: %s, Image: %s",
	// 	podObject.Name,
	// 	podObject.Namespace,
	// 	podObject.Spec.Containers[0].Image,
	// )

	for _, container := range podObject.Spec.Containers {
		log.Printf("Container Name: %s, Image: %s", container.Name, container.Image)
		log.Printf("Namespace: %s", podObject.Namespace)
	}

	allowedNamespace := slices.Contains(dockerHubAllowedNamespaces, podObject.Namespace)
	imageFromDockerHub := strings.HasPrefix(podObject.Spec.Containers[0].Image, "docker.io")

	// Namespace | Image | Allowed
	// default | docker.io | true ( track )
	// default | quay.io | true ( ignore )
	// example | docker.io | false ( track )
	// example | quay.io | true ( ignore )
	// Can not use images from Docker Hub in Namespace - example

	if !allowedNamespace && imageFromDockerHub {
		log.Printf("Pod is not in allowed namespace but the image is from Docker Hub")
		isAllowed = false
		Result = &metav1.Status{
			Code:    403, // HTTP status code for success
			Message: fmt.Sprintf("Can not use images from Docker Hub in Namespace - %s", podObject.Namespace),
		}
	}

	// Create the AdmissionResponse object
	admissionResponse = admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID, // Copy the UID from the request to the response
		Allowed: isAllowed,                   // Set to true to allow the request by default
		Result:  Result,                      // Set the result status
	}

	apiresponse := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionResponse, // Set the response field to the AdmissionResponse object
	}
	c.JSON(http.StatusOK, apiresponse)
}
