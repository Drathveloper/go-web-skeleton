package helper

import "github.com/gin-gonic/gin"

const (
	HXRequestHeader = "HX-Request"
	HXTriggerHeader = "HX-Trigger"
	CloseModalEvent = "closeModal"
)

func IsHTMXRequest(c *gin.Context) bool {
	return c.GetHeader(HXRequestHeader) == "true"
}

func TriggerCloseModal(c *gin.Context) {
	c.Header(HXTriggerHeader, CloseModalEvent)
}
